package metric

// Tags is a map of fields to add to a metrics:
type Tags map[string]string

// Reporter is an interface for collecting and instrumenting metrics
type Reporter interface {
	Count(name string, value float64, tags Tags) error
	Gauge(name string, value float64, tags Tags) error
	Histogram(name string, value float64, tags Tags, opts *HistogramOpts) error
	Summary(name string, value float64, tags Tags) error
	Start() error
	Stop() error
	Name() string
}

type HistogramOpts struct {
}
