package module

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
)

var factories []fx.Option

func List() []fx.Option { return factories }
func Register(m fx.Option) {
	xerror.Assert(m == nil, "[m] should not be null")
	factories = append(factories, m)
}
