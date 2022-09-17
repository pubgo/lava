package grpcs

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/grpc_builder"
)

func init() {
	di.Provide(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Grpc: grpc_builder.GetDefaultCfg(),
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
