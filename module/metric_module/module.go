package metric_module

import (
	"github.com/pubgo/lava/core/metric/metric_builder"
	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"
)

func init() {
	inject.Register(fx.Invoke(metric_builder.Builder))
}
