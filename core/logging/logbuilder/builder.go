package logbuilder

import (
	"io"
	"os"
	"time"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return stack.Stack(pc).Short()
	}
}

// New logger
func New(cfg *logging.Config) log.Logger {
	defer recovery.Exit()

	level := zerolog.DebugLevel
	if cfg.Level != "" {
		level = result.Of(zerolog.ParseLevel(cfg.Level)).Expect("log level is invalid")
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

	// 全局log设置
	ee := logger.With().
		Str(logkey.Hostname, running.Hostname).
		Str(logkey.Project, running.Project).
		Str(logkey.Version, running.Version)

	if running.Namespace != "" {
		ee = ee.Str(logkey.Namespace, running.Namespace)
	}

	logger = ee.Logger()
	log.SetLogger(&logger)

	gl := log.New(&logger)
	for _, ext := range logging.List() {
		ext(gl)
	}
	return gl
}

type writer struct {
	io.Writer
}

func (w writer) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	if err != nil {
		log.Err(err).Str("raw_json", string(p)).Msg("failed to decode invalid json")
		return
	}

	return
}
