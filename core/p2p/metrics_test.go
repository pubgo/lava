package p2p_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/p2p"
)

func TestMetricsRecorderObserveConnect(t *testing.T) {
	scope := tally.NewTestScope("test", nil)
	rec := p2p.NewMetricsRecorder(scope)
	rec.ObserveConnect(p2p.CandidatePairInfo{
		LocalType:  "host",
		RemoteType: "relay",
	}, 150*time.Millisecond)
	rec.SetActiveConnections(1, 1)

	snap := scope.Snapshot()
	assert.NotEmpty(t, snap.Counters)
	assert.NotEmpty(t, snap.Timers)
	assert.NotEmpty(t, snap.Gauges)
}

func TestMetricsRecorderNilScope(t *testing.T) {
	var rec *p2p.MetricsRecorder
	rec.ObserveConnect(p2p.CandidatePairInfo{}, time.Second)
	rec.ObserveDialFailure()
	rec.SetActiveConnections(0, 0)
}

func TestNewMetricsRecorderNil(t *testing.T) {
	assert.Nil(t, p2p.NewMetricsRecorder(nil))
}
