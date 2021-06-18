package exp

import (
	"expvar"

	"github.com/pubgo/lug/metric"
	"github.com/pubgo/xerror"
)

var _ metric.Reporter = (*reporterMetric)(nil)

type reporterMetric struct {
	counters   map[string]*expvar.Float
	gauges     map[string]*expvar.Float
	histograms map[string]*histogram
	summaries  map[string]*histogram
}

func (r *reporterMetric) CreateGauge(name string, labels []string, opts metric.GaugeOpts) (err error) {
	defer xerror.RespErr(&err)

	r.gauges[name] = expvar.NewFloat(name)
	return nil
}

func (r *reporterMetric) CreateCounter(name string, labels []string, opts metric.CounterOpts) (err error) {
	defer xerror.RespErr(&err)

	r.counters[name] = expvar.NewFloat(name)
	return nil
}

func (r *reporterMetric) CreateSummary(name string, labels []string, opts metric.SummaryOpts) (err error) {
	defer xerror.RespErr(&err)

	r.summaries[name] = &histogram{}
	return nil
}

func (r *reporterMetric) CreateHistogram(name string, labels []string, opts metric.HistogramOpts) (err error) {
	defer xerror.RespErr(&err)

	r.histograms[name] = &histogram{}
	return nil
}

func (r *reporterMetric) Count(name string, value float64, tags metric.Tags) error {
	r.counters[name].Set(value)
	return nil
}

func (r *reporterMetric) Gauge(name string, value float64, tags metric.Tags) error {
	r.gauges[name].Set(value)
	return nil
}

func (r *reporterMetric) Histogram(name string, value float64, tags metric.Tags) error {
	return nil
}

func (r *reporterMetric) Summary(name string, value float64, tags metric.Tags) error {
	return nil
}
