package grpcs

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/fiber_builder"
	"github.com/pubgo/lava/internal/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Cfg struct {
	Grpc       *grpc_builder.Cfg  `yaml:"grpc-cfg"`
	Api        *fiber_builder.Cfg `yaml:"rest-cfg"`
	PrintRoute bool               `yaml:"print-route"`
}

func init() {
	dix.Provider(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Grpc: grpc_builder.GetDefaultCfg(),
			Api:  &fiber_builder.Cfg{},
		}
		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
