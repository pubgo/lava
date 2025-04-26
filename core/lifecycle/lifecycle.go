package lifecycle

import "context"

type ExecFunc = func(context.Context) error

type Executor struct {
	Exec ExecFunc
}

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStop(f ExecFunc)
	BeforeStop(f ExecFunc)
	AfterStart(f ExecFunc)
	BeforeStart(f ExecFunc)
}

type Getter interface {
	GetAfterStops() []Executor
	GetBeforeStops() []Executor
	GetAfterStarts() []Executor
	GetBeforeStarts() []Executor
}

var (
	_ Lifecycle = (*lifecycleImpl)(nil)
	_ Getter    = (*lifecycleImpl)(nil)
)

type lifecycleImpl struct {
	beforeStarts []Executor
	afterStarts  []Executor
	beforeStops  []Executor
	afterStops   []Executor
}

func (t *lifecycleImpl) GetAfterStops() []Executor   { return t.afterStops }
func (t *lifecycleImpl) GetBeforeStops() []Executor  { return t.beforeStops }
func (t *lifecycleImpl) GetAfterStarts() []Executor  { return t.afterStarts }
func (t *lifecycleImpl) GetBeforeStarts() []Executor { return t.beforeStarts }
func (t *lifecycleImpl) BeforeStart(f ExecFunc) {
	t.beforeStarts = append(t.beforeStarts, Executor{Exec: f})
}

func (t *lifecycleImpl) BeforeStop(f ExecFunc) {
	t.beforeStops = append([]Executor{{Exec: f}}, t.beforeStops...)
}

func (t *lifecycleImpl) AfterStart(f ExecFunc) {
	t.afterStarts = append(t.afterStarts, Executor{Exec: f})
}

func (t *lifecycleImpl) AfterStop(f ExecFunc) {
	t.afterStops = append([]Executor{{Exec: f}}, t.afterStops...)
}
