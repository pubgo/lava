package slog

import (
	"log/slog"

	"github.com/pubgo/lava/v2/core/logging"

	"github.com/pubgo/funk/v2/log"
)

func init() {
	logging.Register("slog", SetLogger)
}

func SetLogger(logger log.Logger) {
	slog.SetDefault(slog.New(log.NewSlog(logger.WithName("slog"))))
}
