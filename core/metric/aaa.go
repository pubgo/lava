package metric

import (
	"github.com/pubgo/funk/log"
	"github.com/uber-go/tally/v4"
)

type Factory func(cfg *Cfg, log log.Logger) *tally.ScopeOptions
type Tags = map[string]string
type Counter = tally.Counter
type Gauge = tally.Gauge
type Timer = tally.Timer
type Histogram = tally.Histogram
type Capabilities = tally.Capabilities
type Scope = tally.Scope
type Metric = tally.Scope
type Stats = tally.Scope
type Buckets = tally.Buckets
type BucketPair = tally.BucketPair
type Stopwatch = tally.Stopwatch
type StopwatchRecorder = tally.StopwatchRecorder
