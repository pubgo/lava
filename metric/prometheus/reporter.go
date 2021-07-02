package prometheus

import (
	"fmt"
	debug2 "github.com/pubgo/lug/internal/debug"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pubgo/lug/metric"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(metric.Register(Name, New))
}

var _ metric.Reporter = (*reporterMetric)(nil)

//reporterMetric is a prom exporter for go chassis
type reporterMetric struct {
	sync.RWMutex
	registry   prometheus.Registerer
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	summaries  map[string]*prometheus.SummaryVec
	histograms map[string]*prometheus.HistogramVec
}

//New create a prometheus exporter
func New(cfgMap map[string]interface{}) (metric.Reporter, error) {
	var cfg = GetDefaultCfg()
	xerror.Panic(merge.MapStruct(&cfg, cfgMap))
	var reporter = xerror.PanicErr(cfg.Build()).(*reporterMetric)

	debug2.On(func(app *chi.Mux) {
		app.Handle(cfg.Path, promhttp.HandlerFor(prometheus.DefaultGatherer,
			promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	})

	return reporter, nil
}

//CreateGauge create collector
func (c *reporterMetric) CreateGauge(name string, labels []string, opts metric.GaugeOpts) error {
	c.Lock()
	defer c.Unlock()

	_, ok := c.gauges[name]
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", name)
	}

	gVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: opts.Help}, labels)
	c.gauges[name] = gVec
	c.registry.MustRegister(gVec)
	return nil
}

//Gauge set value
func (c *reporterMetric) Gauge(name string, val float64, labels metric.Tags) error {
	c.RLock()
	gVec, ok := c.gauges[name]
	c.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	gVec.With(prometheus.Labels(labels)).Set(val)
	return nil
}

//CreateCounter create collector
func (c *reporterMetric) CreateCounter(name string, labels []string, opts metric.CounterOpts) error {
	c.Lock()
	defer c.Unlock()

	_, ok := c.counters[name]
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", name)
	}

	v := prometheus.NewCounterVec(prometheus.CounterOpts{Name: name, Help: opts.Help}, labels)
	c.counters[name] = v
	c.registry.MustRegister(v)
	return nil
}

//Count increase value
func (c *reporterMetric) Count(name string, val float64, labels metric.Tags) error {
	c.RLock()
	v, ok := c.counters[name]
	c.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Add(val)
	return nil
}

//CreateSummary create collector
func (c *reporterMetric) CreateSummary(name string, labels []string, opts metric.SummaryOpts) error {
	c.Lock()
	defer c.Unlock()

	_, ok := c.summaries[name]
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", name)
	}

	v := prometheus.NewSummaryVec(prometheus.SummaryOpts{Name: name, Help: opts.Help, Objectives: opts.Objectives}, labels)
	c.summaries[name] = v
	c.registry.MustRegister(v)
	return nil
}

//Summary set value
func (c *reporterMetric) Summary(name string, val float64, labels metric.Tags) error {
	c.RLock()
	v, ok := c.summaries[name]
	c.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Observe(val)
	return nil
}

//CreateHistogram create collector
func (c *reporterMetric) CreateHistogram(name string, labels []string, opts metric.HistogramOpts) error {
	c.Lock()
	defer c.Unlock()

	_, ok := c.histograms[name]
	if ok {
		return fmt.Errorf("metric [%s] is duplicated", name)
	}

	v := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: name, Help: opts.Help, Buckets: opts.Buckets}, labels)
	c.histograms[name] = v
	c.registry.MustRegister(v)
	return nil
}

//Histogram set value
func (c *reporterMetric) Histogram(name string, val float64, labels metric.Tags) error {
	c.RLock()
	v, ok := c.histograms[name]
	c.RUnlock()
	if !ok {
		return fmt.Errorf("metrics do not exists, create it first")
	}

	v.With(prometheus.Labels(labels)).Observe(val)
	return nil
}
