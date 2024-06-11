package metrics

import (
	"github.com/pubgo/funk/log"
	tally "github.com/uber-go/tally/v4"
)

type (
	Factory           func(cfg *Config, log log.Logger) *tally.ScopeOptions
	Tags              = map[string]string
	Counter           = tally.Counter
	Gauge             = tally.Gauge
	Timer             = tally.Timer
	Histogram         = tally.Histogram
	Capabilities      = tally.Capabilities
	Scope             = tally.Scope
	Metric            = tally.Scope
	Stats             = tally.Scope
	Buckets           = tally.Buckets
	BucketPair        = tally.BucketPair
	Stopwatch         = tally.Stopwatch
	StopwatchRecorder = tally.StopwatchRecorder
)
