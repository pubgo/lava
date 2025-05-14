package lifecyclebuilder

import "github.com/pubgo/lava/core/lifecycle"

type Provider struct {
	Setter lifecycle.Lifecycle
	Getter lifecycle.Getter
}

func New(handlers []lifecycle.Handler) Provider {
	lc := new(lifecycleImpl)
	for i := range handlers {
		handlers[i](lc)
	}

	return Provider{
		Setter: lc,
		Getter: lc,
	}
}

var (
	_ lifecycle.Lifecycle = (*lifecycleImpl)(nil)
	_ lifecycle.Getter    = (*lifecycleImpl)(nil)
)

type lifecycleImpl struct {
	beforeStarts []lifecycle.Executor
	afterStarts  []lifecycle.Executor
	beforeStops  []lifecycle.Executor
	afterStops   []lifecycle.Executor
}

func (t *lifecycleImpl) GetAfterStops() []lifecycle.Executor   { return t.afterStops }
func (t *lifecycleImpl) GetBeforeStops() []lifecycle.Executor  { return t.beforeStops }
func (t *lifecycleImpl) GetAfterStarts() []lifecycle.Executor  { return t.afterStarts }
func (t *lifecycleImpl) GetBeforeStarts() []lifecycle.Executor { return t.beforeStarts }
func (t *lifecycleImpl) BeforeStart(f lifecycle.ExecFunc) {
	t.beforeStarts = append(t.beforeStarts, lifecycle.Executor{Exec: f})
}

func (t *lifecycleImpl) BeforeStop(f lifecycle.ExecFunc) {
	t.beforeStops = append([]lifecycle.Executor{{Exec: f}}, t.beforeStops...)
}

func (t *lifecycleImpl) AfterStart(f lifecycle.ExecFunc) {
	t.afterStarts = append(t.afterStarts, lifecycle.Executor{Exec: f})
}

func (t *lifecycleImpl) AfterStop(f lifecycle.ExecFunc) {
	t.afterStops = append([]lifecycle.Executor{{Exec: f}}, t.afterStops...)
}
