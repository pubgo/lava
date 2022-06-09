package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/version"
)

func init() {
	tracing.RegisterFactory(Name, func(cfgMap config.CfgMap) error {
		tracing.GetSpanID = GetSpanID

		var cfg = DefaultCfg()
		cfg.ServiceName = runmode.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: logkey.Version, Value: version.Version})
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
