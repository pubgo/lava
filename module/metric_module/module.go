package metric_module

import (
	"github.com/pubgo/lava/core/metric/metric_builder"
	"github.com/pubgo/lava/module"
	"go.uber.org/fx"
)

func init() {
	module.Register(fx.Invoke(metric_builder.Builder))
}
