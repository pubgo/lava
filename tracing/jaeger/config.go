package jaeger

import (
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

type Cfg = jaegerCfg.Configuration

const Name = "jaeger"

var _ = jaeger.NewNullReporter()

func GetDefaultCfg() *Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Exit(err)

	cfg.Disabled = false
	cfg.ServiceName = runenv.Project
	xerror.Panic(err)
	return cfg
}
