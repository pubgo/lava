package import_metric

import (
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Invoke(metric.Builder)
}
