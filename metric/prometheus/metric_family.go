package prometheus

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/pubgo/lug/metric"
)

// metricFamily stores our cached metrics:
type metricFamily struct {
	counters           map[string]*prometheus.CounterVec
	gauges             map[string]*prometheus.GaugeVec
	summaries          map[string]*prometheus.SummaryVec
	histograms         map[string]*prometheus.HistogramVec
	defaultLabels      prometheus.Labels
	mu                 sync.Mutex
	prometheusRegistry *prometheus.Registry
	timingObjectives   map[float64]float64
}

// newMetricFamily returns a new metricFamily (useful in case we want to change the structure later):
func (r *Reporter) newMetricFamily() metricFamily {
	// Take quantile thresholds from our pre-defined list:
	timingObjectives := make(map[float64]float64)
	for _, percentile := range r.cfg.Percentiles {
		if quantileThreshold, ok := quantileThresholds[percentile]; ok {
			timingObjectives[percentile] = quantileThreshold
		}
	}

	return metricFamily{
		counters:           make(map[string]*prometheus.CounterVec),
		gauges:             make(map[string]*prometheus.GaugeVec),
		summaries:          make(map[string]*prometheus.SummaryVec),
		histograms:         make(map[string]*prometheus.HistogramVec),
		defaultLabels:      r.convertTags(r.cfg.Tags),
		prometheusRegistry: r.prometheusRegistry,
		timingObjectives:   timingObjectives,
	}
}

// getCounter either gets a counter, or makes a new one:
func (mf *metricFamily) getCounter(name string, tags metric.Tags) *prometheus.CounterVec {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	// See if we already have this counter:
	counter, ok := mf.counters[name]
	if !ok {

		// Make a new counter:
		counter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        name,
				ConstLabels: mf.defaultLabels,
			},
			listTagKeys(tags),
		)

		// Register it and add it to our list:
		mf.prometheusRegistry.MustRegister(counter)
		mf.counters[name] = counter
	}

	return counter
}

// getGauge either gets a gauge, or makes a new one:
func (mf *metricFamily) getGauge(name string, tags metric.Tags) *prometheus.GaugeVec {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	// See if we already have this gauge:
	gauge, ok := mf.gauges[name]
	if !ok {

		// Make a new gauge:
		gauge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        name,
				ConstLabels: mf.defaultLabels,
			},
			listTagKeys(tags),
		)

		// Register it and add it to our list:
		mf.prometheusRegistry.MustRegister(gauge)
		mf.gauges[name] = gauge
	}

	return gauge
}

// getSummary either gets a summary, or makes a new one:
func (mf *metricFamily) getSummary(name string, tags metric.Tags) *prometheus.SummaryVec {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	// See if we already have this summaryVec:
	summaryVec, ok := mf.summaries[name]
	if !ok {

		// Make a new summaryVec:
		summaryVec = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        name,
				ConstLabels: mf.defaultLabels,
				Objectives:  mf.timingObjectives,
			},
			listTagKeys(tags),
		)

		// Register it and add it to our list:
		mf.prometheusRegistry.MustRegister(summaryVec)
		mf.summaries[name] = summaryVec
	}

	return summaryVec
}

// getHistogram either gets a histogram, or makes a new one:
func (mf *metricFamily) getHistogram(name string, tags metric.Tags, opts *metric.HistogramOpts) *prometheus.HistogramVec {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	// See if we already have this histogram:
	histgm, ok := mf.histograms[name]
	if !ok {

		buckets := prometheus.DefBuckets
		if opts != nil && len(opts.Buckets) > 0 {
			buckets = opts.Buckets
		}

		// Make a new timing:
		histgm = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        name,
				ConstLabels: mf.defaultLabels,
				Buckets:     buckets,
			},
			listTagKeys(tags),
		)

		// Register it and add it to our list:
		mf.prometheusRegistry.MustRegister(histgm)
		mf.histograms[name] = histgm
	}

	return histgm
}
