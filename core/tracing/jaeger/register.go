package jaeger

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/version"
)

func init() {
	tracing.RegisterFactory(Name, func(cfgMap config_type.CfgMap) error {
		tracing.GetSpanID = GetSpanID

		var cfg = DefaultCfg()
		cfg.ServiceName = runtime.Project
		cfg.Tags = append(cfg.Tags, opentracing.Tag{Key: logkey.Version, Value: version.Version})
		xerror.Panic(cfgMap.Decode(&cfg))
		return New(cfg)
	})
}
