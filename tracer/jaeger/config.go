package jaeger

import (
	"github.com/pubgo/xerror"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

type Cfg = jaegerCfg.Configuration

const Name = "jaeger"

func GetDefaultCfg() *Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Panic(err)
	return cfg
}
