package tracer

import (
	"github.com/pubgo/xerror"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"runtime/debug"
)

type Cfg = jaegerCfg.Configuration

func GetDefaultCfg() *Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Panic(err)
	closer, err := cfg.InitGlobalTracer()


	debug.Stack()

	return cfg
}
