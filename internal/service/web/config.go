package grpcs

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/fiber_builder"
)

const (
	Name               = "web"
	defaultContentType = "application/json"
)

type Cfg struct {
	Api        *fiber_builder.Cfg `yaml:"rest-cfg"`
	PrintRoute bool               `yaml:"print-route"`
}

func init() {
	dix.Register(func(c config.Config) *Cfg {
		var cfg = Cfg{
			Api: &fiber_builder.Cfg{},
		}
		xerror.Panic(c.UnmarshalKey(Name, &cfg))
		return &cfg
	})
}
