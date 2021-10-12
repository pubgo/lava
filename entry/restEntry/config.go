package restEntry

import (
	"github.com/pubgo/lug/logger"
	fb "github.com/pubgo/lug/pkg/builder/fiber"
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
