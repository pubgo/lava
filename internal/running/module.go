package running

import (
	"go.uber.org/fx"

	"github.com/pubgo/lava/inject"
)

type Running interface {
	AfterStops(...func())
	BeforeStops(...func())
	AfterStarts(...func())
	BeforeStarts(...func())
}

type GetRunning interface {
	GetAfterStops() []func()
	GetBeforeStops() []func()
	GetAfterStarts() []func()
	GetBeforeStarts() []func()
}

var _ Running = (*runningImpl)(nil)
var _ GetRunning = (*runningImpl)(nil)

type runningImpl struct {
	beforeStarts []func()
	afterStarts  []func()
	beforeStops  []func()
	afterStops   []func()
}

func (t *runningImpl) GetAfterStops() []func()   { return t.afterStops }
func (t *runningImpl) GetBeforeStops() []func()  { return t.beforeStops }
func (t *runningImpl) GetAfterStarts() []func()  { return t.afterStarts }
func (t *runningImpl) GetBeforeStarts() []func() { return t.beforeStarts }
func (t *runningImpl) BeforeStarts(f ...func())  { t.beforeStarts = append(t.beforeStarts, f...) }
func (t *runningImpl) BeforeStops(f ...func())   { t.beforeStops = append(t.beforeStops, f...) }
func (t *runningImpl) AfterStarts(f ...func())   { t.afterStarts = append(t.afterStarts, f...) }
func (t *runningImpl) AfterStops(f ...func())    { t.afterStops = append(t.afterStops, f...) }

func init() {
	inject.Init(func() {
		impl := new(runningImpl)
		inject.Register(fx.Provide(func() Running { return impl }))
		inject.Register(fx.Provide(func() GetRunning { return impl }))
	})
}
