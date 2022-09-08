package grpcs

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/fiber_builder"
	"github.com/pubgo/lava/pkg/grpc_builder"
)

func init() {
	di.Provide(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Grpc: grpc_builder.GetDefaultCfg(),
			Api:  &fiber_builder.Cfg{},
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
