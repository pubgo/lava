package module

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/consts"
)

var factories []fx.Option

func List() []fx.Option { return factories[:] }
func Register(m fx.Option) {
	xerror.Assert(m == nil, "[m] should not be null")
	factories = append(factories, m)
}

func Name(name string) string {
	if name == consts.KeyDefault {
		name = ""
	}
	return name
}

func Init(opts ...fx.Option) {
	_ = fx.New(opts...)
}
