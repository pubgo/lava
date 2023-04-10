package logging

import (
	"os"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/runmode"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"

	"github.com/pubgo/lava/core/logging/logkey"
)

// New logger
func New(cfg *Config) log.Logger {
	defer recovery.Exit()

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	logger := zerolog.New(writer).Level(level).With().Timestamp().Logger()
	if !cfg.AsJson {
		logger = logger.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = writer
		}))
	}
	zl.Logger = logger

	// 全局log设置
	ee := logger.With().Caller().
		Str(logkey.Hostname, runmode.Hostname).
		Str(logkey.Project, runmode.Project).
		Str(logkey.Version, runmode.Version)

	if runmode.Namespace != "" {
		ee = ee.Str(logkey.Namespace, runmode.Namespace)
	}

	logger = ee.Logger()
	log.SetLogger(&logger)

	gl := log.New(&logger)
	for _, ext := range List() {
		ext(gl)
	}
	return gl
}
