package metric

import (
	"github.com/pubgo/lug/logutil"
	"github.com/pubgo/xerror"
)

var defaultScope Scope

// setDefault 设置全局的scope
func setDefault(scope Scope) {
	xerror.Assert(scope == nil, "[scope] should not be nil")
	defaultScope = scope
}

// Root 获取全局的scope
func Root() Scope {
	xerror.Assert(defaultScope == nil, "please set default scope")
	return defaultScope
}

func NewCounter(name string, tags Tags) Counter           { return Root().Counter(name) }
func NewGauge(name string) Gauge                          { return Root().Gauge(name) }
func NewTimer(name string) Timer                          { return Root().Timer(name) }
func NewHistogram(name string, buckets Buckets) Histogram { return Root().Histogram(name, buckets) }
func WithTagged(tags Tags) Scope                          { return Root().Tagged(tags) }
func WithSubScope(name string) Scope                      { return Root().SubScope(name) }

func TimeRecord(t Timer, fn func()) {
	defer xerror.Resp(func(err xerror.XErr) { logutil.ErrLog(err) })

	var start = t.Start()
	fn()
	start.Stop()
}
