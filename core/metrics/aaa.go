// Package metrics 提供基于 uber-go/tally 的指标抽象与 Prometheus 驱动。
//
// 通过 metricbuilder 在 DI 启动时初始化全局 Scope，
// 各模块通过 metrics 包提供的 Counter/Gauge/Timer 等类型上报指标。
package metrics

import (
	"github.com/pubgo/funk/v2/log"
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
