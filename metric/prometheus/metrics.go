package prometheus

import (
	"errors"
	"fmt"

	"github.com/pubgo/golug/metric"
)

var MetricStartError = errors.New("metric start error")

func errHandler(err *error) {
	if r := recover(); r != nil {
		switch r := r.(type) {
		case error:
			*err = r
		default:
			*err = fmt.Errorf("%v", r)
		}
	}
}

// Count is a counter with key/value tags:
// newReporter values are added to any previous one (eg "number of hits")
func (r *Reporter) Count(name string, value float64, tags metric.Tags) (err error) {
	if !r.isStarted.Load() {
		return MetricStartError
	}

	defer errHandler(&err)

	name = r.cfg.Name + name

	counter := r.metrics.getCounter(metric.StripUnsupportedCharacters(name), tags)
	m, _err := counter.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Add(value)
	return
}

// Gauge is a register with key/value tags:
// newReporter values simply override any previous one (eg "current connections")
func (r *Reporter) Gauge(name string, value float64, tags metric.Tags) (err error) {
	if !r.isStarted.Load() {
		return MetricStartError
	}

	defer errHandler(&err)

	name = r.cfg.Name + name

	gauge := r.metrics.getGauge(metric.StripUnsupportedCharacters(name), tags)
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
	if !r.isStarted.Load() {
		return MetricStartError
	}

	defer errHandler(&err)

	name = r.cfg.Name + name

	timing := r.metrics.getSummary(metric.StripUnsupportedCharacters(name), tags)
	m, _err := timing.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Observe(value)
	return
}

// Histogram ...
func (r *Reporter) Histogram(name string, value float64, tags metric.Tags) (err error) {
	if !r.isStarted.Load() {
		return MetricStartError
	}

	defer errHandler(&err)

	name = r.cfg.Name + name

	hm := r.metrics.getHistogram(metric.StripUnsupportedCharacters(name), tags)
	m, _err := hm.GetMetricWith(r.convertTags(tags))
	if _err != nil {
		return _err
	}

	m.Observe(value)
	return
}
