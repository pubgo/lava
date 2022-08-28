package lifecycle

import (
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
)

type executor struct {
	Handler func()
}

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStop(f func())
	BeforeStop(f func())
	AfterStart(f func())
	BeforeStart(f func())
}

type GetLifecycle interface {
	GetAfterStops() []executor
	GetBeforeStops() []executor
	GetAfterStarts() []executor
	GetBeforeStarts() []executor
}

var _ Lifecycle = (*lifecycleImpl)(nil)
var _ GetLifecycle = (*lifecycleImpl)(nil)

type lifecycleImpl struct {
	beforeStarts []executor
	afterStarts  []executor
	beforeStops  []executor
	afterStops   []executor
}

func (t *lifecycleImpl) GetAfterStops() []executor   { return t.afterStops }
func (t *lifecycleImpl) GetBeforeStops() []executor  { return t.beforeStops }
func (t *lifecycleImpl) GetAfterStarts() []executor  { return t.afterStarts }
func (t *lifecycleImpl) GetBeforeStarts() []executor { return t.beforeStarts }
func (t *lifecycleImpl) BeforeStart(f func()) {
	t.beforeStarts = append(t.beforeStarts, executor{Handler: f})
}
func (t *lifecycleImpl) BeforeStop(f func()) {
	t.beforeStops = append(t.beforeStops, executor{Handler: f})
}
func (t *lifecycleImpl) AfterStart(f func()) {
	t.afterStarts = append(t.afterStarts, executor{Handler: f})
}
func (t *lifecycleImpl) AfterStop(f func()) {
	t.afterStops = append(t.afterStops, executor{Handler: f})
}

func New() *lifecycleImpl {
	return new(lifecycleImpl)
}

func init() {
	defer recovery.Exit()

	var lc = new(lifecycleImpl)
	di.Provide(func() Handler { return func(lc Lifecycle) {} })
	di.Provide(func() GetLifecycle { return lc })
	di.Provide(func(handlers []Handler) Lifecycle {
		for i := range handlers {
			handlers[i](lc)
		}
		return lc
	})
}
