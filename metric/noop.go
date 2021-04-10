package metric

import "github.com/pubgo/xerror"

var _ Reporter = (*noopReporter)(nil)

func init() {
	xerror.Exit(Register("noop", func(_ map[string]interface{}) (Reporter, error) {
		return &noopReporter{}, nil
	}))
}

type noopReporter struct{}

func (n *noopReporter) CreateGauge(opts GaugeOpts) error {
	return nil
}

func (n *noopReporter) CreateCounter(opts CounterOpts) error {
	return nil
}

func (n *noopReporter) CreateSummary(opts SummaryOpts) error {
	return nil
}

func (n *noopReporter) CreateHistogram(opts HistogramOpts) error {
	return nil
}

func (n *noopReporter) Count(name string, value float64, tags Tags) error {
	return nil
}

func (n *noopReporter) Gauge(name string, value float64, tags Tags) error {
	return nil
}

func (n *noopReporter) Histogram(name string, value float64, tags Tags) error {
	return nil
}

func (n *noopReporter) Summary(name string, value float64, tags Tags) error {
	return nil
}
