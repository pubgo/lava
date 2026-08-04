package metricbuilder

import (
	"time"

	"github.com/pubgo/funk/v2/typex"
	"github.com/pubgo/funk/v2/vars"
	"github.com/uber-go/tally/v4"

	"github.com/pubgo/lava/v2/core/metrics"
)

func registerVars(m metrics.Metric) {
	vars.Register(vars.UniqueName(metrics.Name, "capabilities"), func() any {
		c := m.Capabilities()
		return typex.Ctx{
			"reporting": c.Reporting(),
			"tagging":   c.Tagging(),
		}
	})

	vars.Register(vars.UniqueName(metrics.Name, "snapshot"), func() any {
		ts, ok := m.(tally.TestScope)
		if !ok {
			return nil
		}
		return serializeSnapshot(ts.Snapshot())
	})
}

func serializeSnapshot(snap tally.Snapshot) map[string]any {
	if snap == nil {
		return nil
	}

	counters := make([]map[string]any, 0, len(snap.Counters()))
	for _, c := range snap.Counters() {
		counters = append(counters, map[string]any{
			"name":  c.Name(),
			"tags":  c.Tags(),
			"value": c.Value(),
		})
	}

	gauges := make([]map[string]any, 0, len(snap.Gauges()))
	for _, g := range snap.Gauges() {
		gauges = append(gauges, map[string]any{
			"name":  g.Name(),
			"tags":  g.Tags(),
			"value": g.Value(),
		})
	}

	timers := make([]map[string]any, 0, len(snap.Timers()))
	for _, tm := range snap.Timers() {
		vals := tm.Values()
		ms := make([]float64, len(vals))
		for i, d := range vals {
			ms[i] = float64(d) / float64(time.Millisecond)
		}
		timers = append(timers, map[string]any{
			"name":   tm.Name(),
			"tags":   tm.Tags(),
			"values": ms, // milliseconds
		})
	}

	histograms := make([]map[string]any, 0, len(snap.Histograms()))
	for _, h := range snap.Histograms() {
		durations := make(map[string]int64, len(h.Durations()))
		for bound, count := range h.Durations() {
			durations[bound.String()] = count
		}
		histograms = append(histograms, map[string]any{
			"name":      h.Name(),
			"tags":      h.Tags(),
			"values":    h.Values(),
			"durations": durations,
		})
	}

	return map[string]any{
		"counters":   counters,
		"gauges":     gauges,
		"timers":     timers,
		"histograms": histograms,
	}
}
