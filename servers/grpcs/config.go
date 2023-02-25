package grpcs

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/config"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/internal/grpc_builder"
)

const (
	Name               = "service"
	defaultContentType = "application/grpc"
)

type Config struct {
	PathPrefix string               `yaml:"path-prefix"`
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	PrintRoute bool                 `yaml:"print-route"`
}

func init() {
	di.Provide(func(c config.Config) *Config {
		var cfg = Config{
			Grpc:       grpc_builder.GetDefaultCfg(),
			PrintRoute: true,
			PathPrefix: version.Project(),
		}

		assert.Must(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
