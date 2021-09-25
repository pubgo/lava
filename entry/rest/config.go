package rest

import (
	fb "github.com/pubgo/lug/builder/fiber"
	"github.com/pubgo/lug/logger"
	"go.uber.org/zap"
)

const Name = "rest_entry"

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named(Name) })
}

type Cfg struct {
	fb.Cfg
}
