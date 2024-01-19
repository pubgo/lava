package logging

import (
	"fmt"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
	"os"
	"time"

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
		logger = logger.Output(&writer{
			ConsoleWriter: zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
				w.Out = os.Stdout
				w.TimeFormat = time.RFC3339
			}),
		})
	}

	// 全局log设置
	ee := logger.With().
		Str(logkey.Hostname, running.Hostname).
		Str(logkey.Project, running.Project).
		Str(logkey.Version, running.Version)

	if running.Namespace != "" {
		ee = ee.Str(logkey.Namespace, running.Namespace)
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

type writer struct {
	zerolog.ConsoleWriter
}

func (w writer) Write(p []byte) (n int, err error) {
	n, err = w.ConsoleWriter.Write(p)
	if err != nil {
		fmt.Println("invalid json: ", string(p))
	}
	return
}
