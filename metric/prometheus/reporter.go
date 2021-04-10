package prometheus

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pubgo/lug/metric"
	"github.com/pubgo/lug/mux"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Exit(metric.Register(Name, func(cfgMap map[string]interface{}) (metric.Reporter, error) {
		var cfg = GetDefaultCfg()
		xerror.Panic(merge.MapStruct(&cfg, cfgMap))
		return NewReporter(cfg)
	}))
}

//reporterMetric is a prom exporter for go chassis
type reporterMetric struct {
	registry   *prometheus.Registry
	lc         sync.RWMutex
	lg         sync.RWMutex
	ls         sync.RWMutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	summaries  map[string]*prometheus.SummaryVec
	histograms map[string]*prometheus.HistogramVec
}

//NewReporter create a prometheus exporter
func NewReporter(cfg Cfg) (metric.Reporter, error) {
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
	registry := prometheus.NewRegistry()
	if cfg.EnableGoRuntimeMetrics {
		xerror.Panic(registry.Register(prometheus.NewGoCollector()))
		xerror.Panic(registry.Register(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{Namespace: "go"})))
		xlog.Info("go runtime metrics is exported")
	}

	mux.On(func(app *chi.Mux) {
		app.Handle(cfg.Path, promhttp.HandlerFor(registry,
			promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	})

	return &reporterMetric{
		registry:   registry,
		lc:         sync.RWMutex{},
		lg:         sync.RWMutex{},
		ls:         sync.RWMutex{},
		summaries:  make(map[string]*prometheus.SummaryVec),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}, nil
}

//CreateGauge create collector
func (c *reporterMetric) CreateGauge(opts metric.GaugeOpts) error {
	c.lg.RLock()
	_, ok := c.gauges[opts.Name]
	c.lg.RUnlock()
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.lg.Lock()
	defer c.lg.Unlock()
	gVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)
	c.gauges[opts.Name] = gVec
	c.registry.MustRegister(gVec)
	return nil
}

//GaugeSet set value
func (c *reporterMetric) Gauge(name string, val float64, labels metric.Tags) error {
	c.lg.RLock()
	gVec, ok := c.gauges[name]
	c.lg.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	gVec.With(prometheus.Labels(labels)).Set(val)
	return nil
}

//CreateCounter create collector
func (c *reporterMetric) CreateCounter(opts metric.CounterOpts) error {
	c.lc.RLock()
	_, ok := c.counters[opts.Name]
	c.lc.RUnlock()
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.lc.Lock()
	defer c.lc.Unlock()
	v := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: opts.Name,
		Help: opts.Help,
	}, opts.Labels)
	c.counters[opts.Name] = v
	c.registry.MustRegister(v)
	return nil
}

//CounterAdd increase value
func (c *reporterMetric) Count(name string, val float64, labels metric.Tags) error {
	c.lc.RLock()
	v, ok := c.counters[name]
	c.lc.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Add(val)
	return nil
}

//CreateSummary create collector
func (c *reporterMetric) CreateSummary(opts metric.SummaryOpts) error {
	c.ls.RLock()
	_, ok := c.summaries[opts.Name]
	c.ls.RUnlock()
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.ls.Lock()
	defer c.ls.Unlock()
	v := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: opts.Objectives,
	}, opts.Labels)
	c.summaries[opts.Name] = v
	c.registry.MustRegister(v)
	return nil
}

//SummaryObserve set value
func (c *reporterMetric) Summary(name string, val float64, labels metric.Tags) error {
	c.ls.RLock()
	v, ok := c.summaries[name]
	c.ls.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Observe(val)
	return nil
}

//CreateHistogram create collector
func (c *reporterMetric) CreateHistogram(opts metric.HistogramOpts) error {
	c.ls.RLock()
	_, ok := c.histograms[opts.Name]
	c.ls.RUnlock()
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", opts.Name)
	}

	c.ls.Lock()
	defer c.ls.Unlock()
	v := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    opts.Name,
		Help:    opts.Help,
		Buckets: opts.Buckets,
	}, opts.Labels)
	c.histograms[opts.Name] = v
	c.registry.MustRegister(v)
	return nil
}

//HistogramObserve set value
func (c *reporterMetric) Histogram(name string, val float64, labels metric.Tags) error {
	c.ls.RLock()
	v, ok := c.histograms[name]
	c.ls.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Observe(val)
	return nil
}
