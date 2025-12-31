package logbuilder

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/log/logfields"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/running"
	"github.com/rs/zerolog"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/core/logging"
	"github.com/pubgo/lava/v2/core/logging/logkey"
)

var GlobalHook zerolog.Hook

// New logger
func New(cfg *logging.Config, hooks []zerolog.Hook) log.Logger {
	defer recovery.Exit(func(err error) error {
		pretty.Println(cfg)
		return err
	})

	level := zerolog.DebugLevel
	if cfg.Level != "" {
		level = result.Wrap(zerolog.ParseLevel(cfg.Level)).Expect("log level is invalid")
	}
	zerolog.SetGlobalLevel(level)

	logger := zerolog.New(&writer{os.Stdout}).Level(level).With().Timestamp().Caller().Logger()
	if !cfg.AsJson {
		logger = logger.Output(&writer{
			zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
				w.Out = os.Stdout
				w.TimeFormat = time.RFC3339
			}),
		})
	}

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
	if len(filters) > 0 {
		expCode := strings.Join(filters, " && ")
		log.Info().Str(logfields.Msg, "log filter expr").Msg(expCode)

		exp := exprFilter(expCode)
		log.SetEnableChecker(func(ctx context.Context, lvl log.Level, name, message string, fields log.Fields) bool {
			envData := map[string]any{"level": lvl.String(), "msg": message, "name": name, "fields": fields}
			output, err := expr.Run(exp, envData)
			if err != nil {
				log.Err(err).Str("expr", expCode).Msg("failed to run log filter expr")
			}
			return err == nil && output.(bool)
		})
	}

	log.SetLogger(lo.ToPtr(ee.Logger()))

	gl := log.GetLogger("ext")
	for _, ext := range logging.List() {
		ext(gl)
	}

	return log.GetLogger()
}

type writer struct {
	io.Writer
}

func (w writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		log.Err(err).Str("raw_json", string(p)).Msg("failed to decode invalid json")
		return n, err
	}

	return n, err
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
