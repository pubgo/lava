package metric

type Handler func(value float64, tags Tags) error

func (m Handler) Record(value float64, tags Tags) error { return m(value, tags) }

// Tags is a map of fields to add to a metrics
type Tags map[string]string

// Reporter is an interface for collecting and instrumenting metrics
type Reporter interface {
	CreateGauge(name string, labels []string, opts GaugeOpts) error
	CreateCounter(name string, labels []string, opts CounterOpts) error
	CreateSummary(name string, labels []string, opts SummaryOpts) error
	CreateHistogram(name string, labels []string, opts HistogramOpts) error

	Count(name string, value float64, tags Tags) error
	Gauge(name string, value float64, tags Tags) error
	Histogram(name string, value float64, tags Tags) error
	Summary(name string, value float64, tags Tags) error
}

//CounterOpts is options to create a counter options
type CounterOpts struct {
	Help string
}

//GaugeOpts is options to create a gauge collector
type GaugeOpts struct {
	Help string
}

//SummaryOpts is options to create summary collector
type SummaryOpts struct {
	Help       string
	Objectives map[float64]float64
}

//HistogramOpts is options to create histogram collector
type HistogramOpts struct {
	Help    string
	Buckets []float64
}
