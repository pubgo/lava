package expvar

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

func (r reporterMetric) CreateGauge(opts metric.GaugeOpts) (err error) {
	defer xerror.RespErr(&err)

	r.gauges[opts.Name] = expvar.NewFloat(opts.Name)
	return nil
}

func (r reporterMetric) CreateCounter(opts metric.CounterOpts) (err error) {
	defer xerror.RespErr(&err)

	r.gauges[opts.Name] = expvar.NewFloat(opts.Name)
	return nil
}

func (r reporterMetric) CreateSummary(opts metric.SummaryOpts) (err error) {
	defer xerror.RespErr(&err)

	r.summaries[opts.Name] = &histogram{}
	return nil
}

func (r reporterMetric) CreateHistogram(opts metric.HistogramOpts) (err error) {
	defer xerror.RespErr(&err)

	r.summaries[opts.Name] = &histogram{}
	return nil
}

func (r reporterMetric) Count(name string, value float64, tags metric.Tags) error {
	panic("implement me")
}

func (r reporterMetric) Gauge(name string, value float64, tags metric.Tags) error {
	panic("implement me")
}

func (r reporterMetric) Histogram(name string, value float64, tags metric.Tags) error {
	panic("implement me")
}

func (r reporterMetric) Summary(name string, value float64, tags metric.Tags) error {
	panic("implement me")
}
