package debug

import (
	cc "github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/builder/chi"
	"github.com/pubgo/lug/logger"
	"go.uber.org/zap"
)

const Name = "debug"

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named(Name) })
}

var Addr = ":8081"

type Mux struct {
	*cc.Mux
	chi.Cfg
	srv chi.Builder
}
