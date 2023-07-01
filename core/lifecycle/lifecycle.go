package lifecycle

type Executor struct {
	Handler func()
}

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStop(f func())
	BeforeStop(f func())
	AfterStart(f func())
	BeforeStart(f func())
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
func (t *lifecycleImpl) BeforeStart(f func()) {
	t.beforeStarts = append(t.beforeStarts, Executor{Handler: f})
}

func (t *lifecycleImpl) BeforeStop(f func()) {
	t.beforeStops = append(t.beforeStops, Executor{Handler: f})
}

func (t *lifecycleImpl) AfterStart(f func()) {
	t.afterStarts = append(t.afterStarts, Executor{Handler: f})
}

func (t *lifecycleImpl) AfterStop(f func()) {
	t.afterStops = append(t.afterStops, Executor{Handler: f})
}
