package metric

import (
	"github.com/pubgo/lava/resource"
	"github.com/uber-go/tally"
	"io"
)

var _ resource.Resource = (*Resource)(nil)

type Resource struct {
	tally.Scope
	io.Closer
}

func (m *Resource) UpdateResObj(val interface{}) { m.Scope = val.(Resource).Scope }
func (m *Resource) Kind() string                 { return Name }

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
