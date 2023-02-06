package grpcs

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/pkg/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type GrpcWebCfg struct {
}

type Cfg struct {
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	GrpcWeb    *GrpcWebCfg          `yaml:"grpc-web"`
	PrintRoute bool                 `yaml:"print-route"`
	BasePrefix string               `yaml:"base-prefix"`
}

func init() {
	di.Provide(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Grpc:       grpc_builder.GetDefaultCfg(),
			PrintRoute: true,
			BasePrefix: version.Project(),
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
