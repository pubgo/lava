package metric

import (
	"github.com/uber-go/tally"
)

type Tags = map[string]string
type Counter = tally.Counter
type Gauge = tally.Gauge
type Timer = tally.Timer
type Histogram = tally.Histogram
type Capabilities = tally.Capabilities
type Scope = tally.Scope
type Buckets = tally.Buckets
type BucketPair = tally.BucketPair
type Stopwatch = tally.Stopwatch
type StopwatchRecorder = tally.StopwatchRecorder
