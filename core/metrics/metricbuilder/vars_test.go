package metricbuilder

import (
	"testing"
	"time"

	"github.com/uber-go/tally/v4"
)

func TestSerializeSnapshot(t *testing.T) {
	scope := tally.NewTestScope("app", nil)
	scope.Counter("hits").Inc(3)
	scope.Gauge("queue").Update(1.5)
	scope.Timer("latency").Record(10 * time.Millisecond)

	out := serializeSnapshot(scope.Snapshot())
	if out == nil {
		t.Fatal("nil snapshot")
	}
	counters, _ := out["counters"].([]map[string]any)
	if len(counters) == 0 {
		t.Fatalf("counters=%v", out["counters"])
	}
	gauges, _ := out["gauges"].([]map[string]any)
	if len(gauges) == 0 {
		t.Fatalf("gauges=%v", out["gauges"])
	}
	timers, _ := out["timers"].([]map[string]any)
	if len(timers) == 0 {
		t.Fatalf("timers=%v", out["timers"])
	}
}

func TestSerializeSnapshot_Nil(t *testing.T) {
	if serializeSnapshot(nil) != nil {
		t.Fatal("expected nil")
	}
}
