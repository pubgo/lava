package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pubgo/lug/metric"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"

	"strings"
)

var Name = "prometheus"
var logs = xlog.GetLogger("metric." + Name)

type Cfg struct {
	Tags                   map[string]string `json:"tags"`
	Percentiles            []float64         `json:"percentiles"`
	Path                   string            `json:"path"`
	Project                string            `json:"project"`
	Name                   string            `json:"name"`
	Prefix                 string            `json:"prefix"`
	EnableGoRuntimeMetrics bool
}

func (cfg Cfg) Build() (_ metric.Reporter, err error) {
	defer xerror.RespErr(&err)

	var name = cfg.Prefix
	name = StripUnsupportedCharacters(strings.ToLower(strings.TrimSpace(name)))
	if name != "" && !strings.HasSuffix(name, "_") {
		name += "_"
	}
	cfg.Prefix = name

	cfg.Path = strings.ToLower(strings.TrimSpace(cfg.Path))
	if cfg.Path == "" {
		cfg.Path = "/metrics"
	}

	// Make a prometheus registry (this keeps track of any metrics we generate):
	registry := prometheus.DefaultRegisterer
	if cfg.EnableGoRuntimeMetrics {
		xerror.Panic(registry.Register(prometheus.NewGoCollector()))
		xerror.Panic(registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{Namespace: "go"})))
		logs.Info("go runtime metrics is exported")
	}

	return &reporterMetric{
		registry:   registry,
		summaries:  make(map[string]*prometheus.SummaryVec),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}, nil
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Path:        "/metrics",
		Percentiles: []float64{0, 0.5, 0.75, 0.90, 0.95, 0.98, 0.99, 1},
		Prefix:      "/registry",
	}
}
