package metric

import (
	"github.com/pubgo/x/attrs"

	"net/http"
)

var Name = "metric"
var cfg = make(map[string]Cfg)

var (
	DefaultServeMux = &http.ServeMux{}
)

type Cfg struct {
	Percentiles []float64        `json:"percentiles"`
	Path        string           `json:"path"`
	Project     string           `json:"project"`
	Driver      string           `json:"driver"`
	Name        string           `json:"name"`
	Prefix      string           `json:"prefix"`
	Attrs       attrs.Attributes `json:"-"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		// This is the endpoint where the Prometheus metrics will be made available ("/metrics" is the default with Prometheus):
		Path: "/metrics",
		// defaultPercentiles is the default spread of percentiles/quantiles we maintain for timings/histogram metrics:
		Percentiles: []float64{0, 0.5, 0.75, 0.90, 0.95, 0.98, 0.99, 1},
		Prefix:      "/registry",
		Driver:      "mdns",
	}
}
