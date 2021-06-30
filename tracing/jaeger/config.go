package jaeger

import (
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

type Cfg = jaegerCfg.Configuration

const Name = "jaeger"

var logs = xlog.GetLogger(Name)

func GetDefaultCfg() *Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Panic(err)
	return cfg
}
