package util

import (
	"github.com/pubgo/lava/logging"
	"github.com/rsocket/rsocket-go/logger"
)

func init() {
	logger.SetLevel(logger.LevelDebug)
	logger.SetLogger(logging.S())
}
