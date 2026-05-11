package logbuilder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/features"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/log/logfields"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/pubgo/lava/v2/core/logging"
	"github.com/pubgo/lava/v2/core/logging/logkey"
	"github.com/pubgo/lava/v2/core/running"
)

var GlobalHook zerolog.Hook

// ConsoleLogEnabled 控制终端日志输出的开关
var ConsoleLogEnabled = features.Bool("log.console.enabled", true, "是否启用终端日志输出")

// New 创建新的 logger 实例
func New(cfg *logging.Config, hooks []zerolog.Hook) log.Logger {
	defer recovery.Exit(func(err error) error {
		pretty.Println(cfg)
		return err
	})

	// 配置验证
	if err := cfg.Validate(); err != nil {
		assert.Exit(err, "invalid logging config")
	}

	// 设置禁用的 loggers
	if len(cfg.DisableLoggers) > 0 {
		logging.SetDisabledLoggers(cfg.DisableLoggers)
	}

	level := zerolog.DebugLevel
	if cfg.Level != "" {
		level = result.Wrap(zerolog.ParseLevel(cfg.Level)).Expect("log level is invalid")
	}
	zerolog.SetGlobalLevel(level)

	// 构建输出 writers
	var writers []io.Writer

	// 终端输出（受 feature flag 控制）
	var consoleWriter io.Writer
	if cfg.AsJson {
		consoleWriter = os.Stdout
	} else {
		consoleWriter = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stdout
			w.TimeFormat = time.RFC3339
		})
	}
	// 包装 consoleWriter，根据 feature flag 动态控制
	writers = append(writers, &consoleWriterWrapper{w: consoleWriter})

	// 文件输出（JSON 格式，带 logrotate）
	if cfg.File != nil && cfg.File.Enabled {
		fileWriter := newFileWriter(cfg.File)
		writers = append(writers, fileWriter)
		// 保存文件路径供 loggerdebug 使用
		logging.SetLogFilePath(cfg.File.Path)
	}

	// 组合多个 writer
	multiWriter := io.MultiWriter(writers...)
	logger := zerolog.New(&writer{multiWriter}).Level(level).With().Timestamp().Caller().Logger()

	if GlobalHook != nil {
		hooks = append(hooks, GlobalHook)
	}
	logger = logger.Hook(hooks...)

	// 全局log设置
	ee := logger.With().
		Str(logkey.Hostname, running.Hostname).
		Str(logkey.Project, running.Project()).
		Str(logkey.Version, running.Version())

	if running.Namespace != "" {
		ee = ee.Str(logkey.Namespace, running.Namespace)
	}

	filters := lo.Filter(cfg.Filters, func(item string, index int) bool { return strings.TrimSpace(item) != "" })

	// 设置日志过滤器
	if len(filters) > 0 || len(cfg.DisableLoggers) > 0 {
		var exp *vm.Program
		var expCode string

		if len(filters) > 0 {
			expCode = strings.Join(filters, " && ")
			log.Info().Str(logfields.Msg, "log filter expr").Msg(expCode)
			exp = exprFilter(expCode)
		}

		log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Fields) bool {
			// 检查 logger name 是否被禁用
			if logging.IsDisabled(name) {
				return false
			}

			// 检查 expr 过滤器
			if exp != nil {
				envData := map[string]any{"level": lvl.String(), "msg": message, "name": name, "fields": fields}
				output, err := expr.Run(exp, envData)
				if err != nil {
					log.Err(err).Str("expr", expCode).Msg("failed to run log filter expr")
					return true // 出错时不过滤
				}
				return output.(bool)
			}

			return true
		})
	}

	log.SetLogger(lo.ToPtr(ee.Logger()))

	// 初始化扩展 loggers
	gl := log.GetLogger("ext")
	for name, ext := range logging.List() {
		log.Info().Str("logger", name).Msg("initializing log extension")
		ext(gl)
	}

	return log.GetLogger()
}

// newFileWriter 创建带 logrotate 的文件 writer
func newFileWriter(cfg *logging.FileConfig) io.Writer {
	// 确保目录存在
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Err(err).Str("dir", dir).Msg("failed to create log directory")
	}

	return &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 保留的旧文件数量
		MaxAge:     cfg.MaxAge,     // 保留天数
		Compress:   cfg.Compress,   // 是否压缩
		LocalTime:  true,           // 使用本地时间
	}
}

type writer struct {
	io.Writer
}

func (w writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		// 使用 stderr 直接输出避免递归
		_, _ = fmt.Fprintf(os.Stderr, "[logging] write error: %v, raw: %s\n", err, string(p))
	}
	return n, err
}

// consoleWriterWrapper 包装 console writer，根据 feature flag 动态控制输出
type consoleWriterWrapper struct {
	w io.Writer
}

func (c *consoleWriterWrapper) Write(p []byte) (n int, err error) {
	if !ConsoleLogEnabled.Value() {
		return len(p), nil // 禁用时丢弃输出
	}
	return c.w.Write(p)
}

func exprFilter(code string) *vm.Program {
	env := map[string]any{"level": "", "name": "", "msg": "", "fields": log.Fields{}}

	program, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		log.Err(err).Str("expr", code).Msg("failed to compile log filter expr")
		assert.Exit(err, code)
	}
	return program
}
