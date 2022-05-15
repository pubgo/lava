package import_metric

import (
	"github.com/pubgo/lava/core/metric/metric_builder"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(metric_builder.Builder)
}
