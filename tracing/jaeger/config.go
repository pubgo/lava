package jaeger

import (
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/xerror"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

type Cfg struct {
	*jaegerCfg.Configuration
	BatchSize int `yaml:"batch_size"`
}

const Name = "jaeger"

func GetDefaultCfg() *Cfg {
	cfg, err := jaegerCfg.FromEnv()
	xerror.Exit(err)

	cfg.Disabled = false
	cfg.ServiceName = runenv.Project
	xerror.Panic(err)
	return &Cfg{Configuration: cfg, BatchSize: 100}
}
