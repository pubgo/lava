package lifecycle

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/lava/internal/pkg/utils"
)

type executor struct {
	Handler func() error
	Msg     string
}

type Handler func(lc Lifecycle)

type Lifecycle interface {
	AfterStop(f func() error, msg ...string)
	BeforeStop(f func() error, msg ...string)
	AfterStart(f func() error, msg ...string)
	BeforeStart(f func() error, msg ...string)
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
func (t *lifecycleImpl) BeforeStart(f func() error, msg ...string) {
	t.beforeStarts = append(t.beforeStarts, executor{
		Handler: f,
		Msg:     utils.FirstNotEmpty(msg...),
	})
}
func (t *lifecycleImpl) BeforeStop(f func() error, msg ...string) {
	t.beforeStops = append(t.beforeStops, executor{
		Handler: f,
		Msg:     utils.FirstNotEmpty(msg...),
	})
}
func (t *lifecycleImpl) AfterStart(f func() error, msg ...string) {
	t.afterStarts = append(t.afterStarts, executor{
		Handler: f,
		Msg:     utils.FirstNotEmpty(msg...),
	})
}
func (t *lifecycleImpl) AfterStop(f func() error, msg ...string) {
	t.afterStops = append(t.afterStops, executor{
		Handler: f,
		Msg:     utils.FirstNotEmpty(msg...),
	})
}

func New() *lifecycleImpl {
	return new(lifecycleImpl)
}

func init() {
	defer recovery.Exit()

	var lc = new(lifecycleImpl)
	dix.Provider(func() Handler { return func(lc Lifecycle) {} })
	dix.Provider(func() GetLifecycle { return lc })
	dix.Provider(func(handlers []Handler) Lifecycle {
		for i := range handlers {
			handlers[i](lc)
		}
		return lc
	})
}
