package prometheus

var Name = "prometheus"

type Cfg struct {
	Tags                   map[string]string `json:"tags"`
	Percentiles            []float64         `json:"percentiles"`
	Path                   string            `json:"path"`
	Project                string            `json:"project"`
	Name                   string            `json:"name"`
	Prefix                 string            `json:"prefix"`
	EnableGoRuntimeMetrics bool
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Path:        "/metrics",
		Percentiles: []float64{0, 0.5, 0.75, 0.90, 0.95, 0.98, 0.99, 1},
		Prefix:      "/registry",
	}
}
