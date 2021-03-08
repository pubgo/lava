package metric

// Tags is a map of fields to add to a metrics:
type Tags map[string]string

// Reporter is an interface for collecting and instrumenting metrics
type Reporter interface {
	Count(name string, value float64, tags Tags) error
	Gauge(name string, value float64, tags Tags) error
	Histogram(name string, value float64, tags Tags) error
	Summary(name string, value float64, tags Tags) error
	Start() error
	Stop() error
	String() string
}

// Counter describes a metrics that accumulates values monotonically.
// An example of a counter is the number of received HTTP requests.
type Counter interface {
	With(tags Tags) Counter
	Add(delta float64) error
}

// Gauger describes a metrics that takes specific values over time.
// An example of a gauge is the current depth of a job queue.
type Gauger interface {
	With(tags Tags) Gauger
	Set(value float64) error
	Add(delta float64) error
}

// Histogram describes a metrics that takes repeated observations of the same
// kind of thing, and produces a statistical summary of those observations,
// typically expressed as buckets. An example of a histogram is
// HTTP request latencies.
type Histogramer interface {
	With(tags Tags) Histogramer
	Observe(value float64) error
}

// Summarier describes a metrics that takes repeated observations of the same
// kind of thing, and produces a statistical summary of those observations,
// typically expressed as quantiles. An example of a summary is
// HTTP request latencies.
type Summarier interface {
	With(tags Tags) Summarier
	Observe(value float64) error
}
