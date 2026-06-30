package lifecyclebuilder

import "github.com/pubgo/lava/v2/core/lifecycle"

// Provider 同时暴露生命周期钩子的写入侧（Setter）与读取侧（Getter），
// 供依赖注入容器装配使用。
type Provider struct {
	Setter lifecycle.Lifecycle
	Getter lifecycle.Getter
}

// New 通过给定的 handlers 构建生命周期 Provider，
// 每个 handler 在构建期向内部实现注册各阶段钩子。
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
// BeforeStart 追加注册（FIFO）：启动前按注册顺序执行。
func (t *lifecycleImpl) BeforeStart(f lifecycle.ExecFunc) {
	t.beforeStarts = append(t.beforeStarts, lifecycle.Executor{Exec: f})
}

// BeforeStop 前插注册（LIFO）：停止前按注册逆序执行。
func (t *lifecycleImpl) BeforeStop(f lifecycle.ExecFunc) {
	t.beforeStops = append([]lifecycle.Executor{{Exec: f}}, t.beforeStops...)
}

// AfterStart 追加注册（FIFO）：启动后按注册顺序执行。
func (t *lifecycleImpl) AfterStart(f lifecycle.ExecFunc) {
	t.afterStarts = append(t.afterStarts, lifecycle.Executor{Exec: f})
}

// AfterStop 前插注册（LIFO）：停止后按注册逆序执行。
func (t *lifecycleImpl) AfterStop(f lifecycle.ExecFunc) {
	t.afterStops = append([]lifecycle.Executor{{Exec: f}}, t.afterStops...)
}
