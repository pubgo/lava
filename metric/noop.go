package metric

import "github.com/pubgo/xerror"

var _ Reporter = (*noopReporter)(nil)

func init() {
	xerror.Exit(Register("noop", newNoopReporter))
}

func newNoopReporter(_ map[string]interface{}) (Reporter, error) { return &noopReporter{}, nil }

type noopReporter struct{}

func (n *noopReporter) CreateGauge(name string, labels []string, opts GaugeOpts) error     { return nil }
func (n *noopReporter) CreateCounter(name string, labels []string, opts CounterOpts) error { return nil }
func (n *noopReporter) CreateSummary(name string, labels []string, opts SummaryOpts) error { return nil }
func (n *noopReporter) CreateHistogram(name string, labels []string, opts HistogramOpts) error {
	return nil
}
func (n *noopReporter) Count(name string, value float64, tags Tags) error     { return nil }
func (n *noopReporter) Gauge(name string, value float64, tags Tags) error     { return nil }
func (n *noopReporter) Histogram(name string, value float64, tags Tags) error { return nil }
func (n *noopReporter) Summary(name string, value float64, tags Tags) error   { return nil }
