package metric

import (
	"go.uber.org/atomic"
)

// counter implements Counter
type counter struct {
	name string
	tags Tags
}

// NewCounter define counter with name
func NewCounter(name string) Counter {
	return &counter{name: name}
}

// With implements Counter.
func (c *counter) With(tags Tags) Counter {
	return &counter{
		name: c.name,
		tags: tags,
	}
}

// Add implements Counter.
func (c *counter) Add(delta float64) error {
	return Count(c.name, delta, c.tags)
}

// gauge implements IGauge
type gauge struct {
	name  string
	tags  Tags
	value atomic.Float64
}

// NewGauge define gauge with name
func NewGauge(name string) Gauger {
	return &gauge{name: name}
}

// With implements IGauge.
func (g *gauge) With(tags Tags) Gauger {
	return &gauge{
		name: g.name,
		tags: tags,
	}
}

// Set implements IGauge.
func (g *gauge) Set(value float64) error {
	g.value.Store(value)
	return Gauge(g.name, value, g.tags)
}

// Add implements IGauge.
func (g *gauge) Add(delta float64) error {
	g.value.Add(delta)
	return Gauge(g.name, g.value.Load(), g.tags)
}

// histogram implements IHistogram
type histogram struct {
	name string
	tags Tags
}

// NewHistogram define histogram with name
func NewHistogram(name string) Histogramer {
	return &histogram{
		name: name,
	}
}

// With implements IHistogram.
func (h *histogram) With(tags Tags) Histogramer {
	return &histogram{
		name: h.name,
		tags: tags,
	}
}

// Observe implements IHistogram.
func (h *histogram) Observe(value float64) error {
	return Histogram(h.name, value, h.tags)
}

// summary implements ISummary.
type summary struct {
	name string
	tags Tags
}

// NewSummary define summary with name
func NewSummary(name string) Summarier {
	return &summary{
		name: name,
	}
}

// With implements ISummary.
func (h *summary) With(tags Tags) Summarier {
	return &summary{
		name: h.name,
		tags: tags,
	}
}

// Observe implements ISummary.
func (h *summary) Observe(value float64) error {
	return Summary(h.name, value, h.tags)
}
