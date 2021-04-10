package metric

// Tags is a map of fields to add to a metrics
type Tags map[string]string
type Factory func(cfg map[string]interface{}) (Reporter, error)

// Reporter is an interface for collecting and instrumenting metrics
type Reporter interface {
	CreateGauge(opts GaugeOpts) error
	CreateCounter(opts CounterOpts) error
	CreateSummary(opts SummaryOpts) error
	CreateHistogram(opts HistogramOpts) error

	Count(name string, value float64, tags Tags) error
	Gauge(name string, value float64, tags Tags) error
	Histogram(name string, value float64, tags Tags) error
	Summary(name string, value float64, tags Tags) error
}

//CounterOpts is options to create a counter options
type CounterOpts struct {
	Name   string
	Help   string
	Labels []string
}

//GaugeOpts is options to create a gauge collector
type GaugeOpts struct {
	Name   string
	Help   string
	Labels []string
}

//SummaryOpts is options to create summary collector
type SummaryOpts struct {
	Name       string
	Help       string
	Labels     []string
	Objectives map[float64]float64
}

//HistogramOpts is options to create histogram collector
type HistogramOpts struct {
	Name    string
	Help    string
	Labels  []string
	Buckets []float64
}
