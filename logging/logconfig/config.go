package logconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/pubgo/lava/pkg/utils"
)

var (
	allLevels = []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel, zapcore.FatalLevel}
)

func GlobalLevel(l ...zapcore.Level) string {
	if globalLevel == nil {
		return ""
	}

	if len(l) > 0 {
		return globalLevel.String()
	}

	globalLevel.SetLevel(l[0])
	return l[0].String()
}

var globalLevel *zap.AtomicLevel

type option func(opts *Config)

type encoderConfig struct {
	MessageKey       string `json:"messageKey" yaml:"messageKey" toml:"messageKey"`
	LevelKey         string `json:"levelKey" yaml:"levelKey" toml:"levelKey"`
	TimeKey          string `json:"timeKey" yaml:"timeKey" toml:"timeKey"`
	NameKey          string `json:"nameKey" yaml:"nameKey" toml:"nameKey"`
	CallerKey        string `json:"callerKey" yaml:"callerKey" toml:"callerKey"`
	StacktraceKey    string `json:"stacktraceKey" yaml:"stacktraceKey" toml:"stacktraceKey"`
	LineEnding       string `json:"lineEnding" yaml:"lineEnding" toml:"lineEnding"`
	EncodeLevel      string `json:"levelEncoder" yaml:"levelEncoder" toml:"levelEncoder"`
	EncodeTime       string `json:"timeEncoder" yaml:"timeEncoder" toml:"timeEncoder"`
	EncodeDuration   string `json:"durationEncoder" yaml:"durationEncoder" toml:"durationEncoder"`
	EncodeCaller     string `json:"callerEncoder" yaml:"callerEncoder" toml:"callerEncoder"`
	EncodeName       string `json:"nameEncoder" yaml:"nameEncoder" toml:"nameEncoder"`
	ConsoleSeparator string `json:"consoleSeparator" yaml:"consoleSeparator"`
}

type samplingConfig struct {
	Initial    int `json:"initial" yaml:"initial" toml:"initial"`
	Thereafter int `json:"thereafter" yaml:"thereafter" toml:"thereafter"`
}

type rotateCfg struct {
	Filename   string `json:"filename" yaml:"filename"`
	MaxSize    int    `json:"maxsize" yaml:"maxsize"`
	MaxAge     int    `json:"maxage" yaml:"maxage"`
	MaxBackups int    `json:"maxbackups" yaml:"maxbackups"`
	LocalTime  bool   `json:"localtime" yaml:"localtime"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

type Config struct {
	Level             string                 `json:"level" yaml:"level" toml:"level"`
	Development       bool                   `json:"development" yaml:"development" toml:"development"`
	DisableCaller     bool                   `json:"disableCaller" yaml:"disableCaller" toml:"disableCaller"`
	DisableStacktrace bool                   `json:"disableStacktrace" yaml:"disableStacktrace" toml:"disableStacktrace"`
	Sampling          *samplingConfig        `json:"sampling" yaml:"sampling" toml:"sampling"`
	Encoding          string                 `json:"encoding" yaml:"encoding" toml:"encoding"`
	EncoderConfig     encoderConfig          `json:"encoderConfig" yaml:"encoderConfig" toml:"encoderConfig"`
	OutputPaths       []string               `json:"outputPaths" yaml:"outputPaths" toml:"outputPaths"`
	ErrorOutputPaths  []string               `json:"errorOutputPaths" yaml:"errorOutputPaths" toml:"errorOutputPaths"`
	InitialFields     map[string]interface{} `json:"initialFields" yaml:"initialFields" toml:"initialFields"`
	Rotate            *rotateCfg             `json:"rotate" yaml:"rotate"`
	SamplingHook      func(zapcore.Entry)    `json:"-" yaml:"-"`
}

func (t Config) handleOpts(opts ...option) Config {
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

func (t Config) Build(name string, opts ...zap.Option) (_ *zap.Logger) {
	zapCfg := zap.Config{}
	var dt = assert.Must1(json.Marshal(&t))
	assert.Must(json.Unmarshal(dt, &zapCfg))

	// 保留全局log level, 用于后期动态修改
	globalLevel = &zapCfg.Level

	key := utils.FirstNotEmpty(t.EncoderConfig.EncodeLevel, defaultKey)
	zapCfg.EncoderConfig.EncodeLevel = levelEncoder[key]

	key = utils.FirstNotEmpty(t.EncoderConfig.EncodeTime, defaultKey)
	zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(key)
	if encoder, ok := timeEncoder[key]; ok {
		zapCfg.EncoderConfig.EncodeTime = encoder
	}

	key = utils.FirstNotEmpty(t.EncoderConfig.EncodeDuration, defaultKey)
	zapCfg.EncoderConfig.EncodeDuration = durationEncoder[key]

	key = utils.FirstNotEmpty(t.EncoderConfig.EncodeCaller, defaultKey)
	zapCfg.EncoderConfig.EncodeCaller = callerEncoder[key]

	key = utils.FirstNotEmpty(t.EncoderConfig.EncodeName, defaultKey)
	zapCfg.EncoderConfig.EncodeName = nameEncoder[key]

	// 采样hook设置
	if t.SamplingHook != nil {
		zapCfg.Sampling.Hook = func(entry zapcore.Entry, decision zapcore.SamplingDecision) {
			if decision == zapcore.LogDropped {
				return
			}

			t.SamplingHook(entry)
		}
	}

	var log = assert.Must1(zapCfg.Build(opts...))

	if t.Rotate != nil {
		var cores []zapcore.Core
		for i := range allLevels {
			lvl := allLevels[i]
			var w = &lumberjack.Logger{
				Filename:   os.ExpandEnv(filepath.Join(t.Rotate.Filename, lvl.String(), fmt.Sprintf("%s.log", name))),
				MaxSize:    t.Rotate.MaxSize,
				MaxBackups: t.Rotate.MaxBackups,
				MaxAge:     t.Rotate.MaxAge,
				Compress:   t.Rotate.Compress,
			}

			var cfg = zapCfg.EncoderConfig
			cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
			cores = append(cores, zapcore.NewCore(
				zapcore.NewJSONEncoder(cfg), zapcore.AddSync(w),
				zap.LevelEnablerFunc(func(level zapcore.Level) bool { return level == lvl })))
		}

		log = log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core { return zapcore.NewTee(append(cores, core)...) }))
	}
	return log
}

func NewDevConfig(opts ...option) Config {
	cfg := Config{
		Level:             "debug",
		Development:       true,
		Encoding:          "console",
		DisableStacktrace: true,
		Rotate: &rotateCfg{
			Filename:   "${cfg_dir}/logs",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   false,
		},
		EncoderConfig: encoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			EncodeLevel:    "capitalColor",
			EncodeTime:     "iso8601",
			EncodeDuration: "string",
			EncodeCaller:   "full",
			LineEnding:     zapcore.DefaultLineEnding,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return cfg.handleOpts(opts...)
}

func NewProdConfig(opts ...option) Config {
	cfg := Config{
		Level:       "info",
		Development: false,
		Sampling: &samplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          "json",
		DisableStacktrace: true,
		EncoderConfig: encoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    "default",
			EncodeTime:     "default",
			EncodeDuration: "default",
			EncodeCaller:   "default",
			LineEnding:     zapcore.DefaultLineEnding,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return cfg.handleOpts(opts...)
}
