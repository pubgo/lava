package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/recovery"
	"go.etcd.io/etcd/api/v3/version"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging/logkey"
)

func init() {
	defer recovery.Exit()

	tracing.RegisterFactory(Name, func(cfgMap config.CfgMap) error {
		tracing.GetSpanID = GetSpanID

		var cfg = DefaultCfg()
		cfg.ServiceName = runmode.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: logkey.Version, Value: version.Version})
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
