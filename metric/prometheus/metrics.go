package prometheus

import (
	"github.com/pubgo/golug/metric"
	"github.com/pubgo/xerror"
)

// Count is a counter with key/value tags:
// newReporter values are added to any previous one (eg "number of hits")
func (r *Reporter) Count(name string, value float64, tags metric.Tags) (gErr error) {
	defer xerror.RespErr(&gErr)

	name = r.cfg.Name + name

	counter := r.metrics.getCounter(StripUnsupportedCharacters(name), tags)
	m, err := counter.GetMetricWith(r.convertTags(tags))
	if err != nil {
		return err
	}

	m.Add(value)
	return
}

// Gauge is a register with key/value tags:
// newReporter values simply override any previous one (eg "current connections")
func (r *Reporter) Gauge(name string, value float64, tags metric.Tags) (err error) {
	defer xerror.RespErr(&err)

	name = r.cfg.Name + name

	gauge := r.metrics.getGauge(StripUnsupportedCharacters(name), tags)
	m, _err := gauge.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Set(value)
	return
}

// Summarier is a histogram with key/valye tags:
// Reporter values are added into a series of aggregations
func (r *Reporter) Summary(name string, value float64, tags metric.Tags) (err error) {
	defer xerror.RespErr(&err)

	name = r.cfg.Name + name

	timing := r.metrics.getSummary(StripUnsupportedCharacters(name), tags)
	m, _err := timing.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Observe(value)
	return
}

// Histogram ...
func (r *Reporter) Histogram(name string, value float64, tags metric.Tags, opts *metric.HistogramOpts) (err error) {
	defer xerror.RespErr(&err)

	name = r.cfg.Name + name

	hm := r.metrics.getHistogram(StripUnsupportedCharacters(name), tags, opts)
	m, _err := hm.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Observe(value)
	return
}
