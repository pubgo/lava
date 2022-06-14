package lifecycle

import "github.com/pubgo/dix"

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
}

type GetLifecycle interface {
	GetAfterStops() []func()
	GetBeforeStops() []func()
	GetAfterStarts() []func()
	GetBeforeStarts() []func()
}

var _ Lifecycle = (*lifecycleImpl)(nil)
var _ GetLifecycle = (*lifecycleImpl)(nil)

type lifecycleImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
}

func (t *lifecycleImpl) GetAfterStops() []func()   { return t.afterStops }
func (t *lifecycleImpl) GetBeforeStops() []func()  { return t.beforeStops }
func (t *lifecycleImpl) GetAfterStarts() []func()  { return t.afterStarts }
func (t *lifecycleImpl) GetBeforeStarts() []func() { return t.beforeStarts }
func (t *lifecycleImpl) BeforeStarts(f ...func())  { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *lifecycleImpl) BeforeStops(f ...func())   { t.beforeStops = append(t.beforeStops, f...) }
func (t *lifecycleImpl) AfterStarts(f ...func())   { t.afterStarts = append(t.afterStarts, f...) }
func (t *lifecycleImpl) AfterStops(f ...func())    { t.afterStops = append(t.afterStops, f...) }

func New() *lifecycleImpl {
	return new(lifecycleImpl)
}

func init() {
	var lc = new(lifecycleImpl)
	dix.Provider(func() Handler { return func(lc Lifecycle) {} })
	dix.Provider(func(handlers []Handler) Lifecycle {
		for i := range handlers {
			handlers[i](lc)
		}
		return lc
	})
	dix.Provider(func() GetLifecycle { return lc })
}
