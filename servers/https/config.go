package https

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/grpc_builder"
	"github.com/pubgo/lava/version"
)

const (
	Name               = "service"
	defaultContentType = "application/json"
)

type Cfg struct {
	Grpc       *grpc_builder.Config `yaml:"grpc-server"`
	PrintRoute bool                 `yaml:"print-route"`
	BasePrefix string               `json:"base-prefix"`
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
