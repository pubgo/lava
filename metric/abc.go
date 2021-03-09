package metric

var Name = "metric"

// Tags is a map of fields to add to a metrics:
type Tags map[string]string
type Factory  func(cfg map[string]interface{}) (Reporter, error)

// Reporter is an interface for collecting and instrumenting metrics
type Reporter interface {
	Count(name string, value float64, tags Tags) error
	Gauge(name string, value float64, tags Tags) error
	Histogram(name string, value float64, tags Tags, opts *HistogramOpts) error
	Summary(name string, value float64, tags Tags) error
}

type HistogramOpts struct {
	Buckets []float64
}
