package slog

import (
	"log/slog"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/v2/core/logging"
)

func init() {
	logging.Register("slog", SetLogger)
}

func SetLogger(logger log.Logger) {
	slog.SetDefault(slog.New(log.NewSlog(logger.WithName("slog"))))
}
