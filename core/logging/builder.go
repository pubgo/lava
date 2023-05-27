package logging

import (
	"os"
	"time"

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
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Caller().Logger()
	if !cfg.AsJson {
		logger = logger.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stdout
			w.TimeFormat = time.RFC3339
		}))
	}

	// 全局log设置
	ee := logger.With().
		Str(logkey.Hostname, runmode.Hostname).
		Str(logkey.Project, runmode.Project).
		Str(logkey.Version, runmode.Version)

	if runmode.Namespace != "" {
		ee = ee.Str(logkey.Namespace, runmode.Namespace)
	}

	logger = ee.Logger()
	zl.Logger = logger
	log.SetLogger(&logger)

	gl := log.New(&logger)
	for _, ext := range List() {
		ext(gl)
	}
	return gl
}
