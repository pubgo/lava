// Package lifecycle 定义服务生命周期钩子的抽象。
//
// 它将启动/停止划分为四个阶段：BeforeStart、AfterStart、BeforeStop、AfterStop。
// 启动阶段（Before/AfterStart）按注册顺序（FIFO）执行；
// 停止阶段（Before/AfterStop）按注册的逆序（LIFO）执行，
// 以保证资源按「后初始化先释放」的顺序安全清理。
package lifecycle

import "context"

// WrapNoError 将一个无返回值的函数适配为 ExecFunc。
func WrapNoError(fn func(context.Context)) ExecFunc {
	return func(ctx context.Context) error { fn(ctx); return nil }
}

// WrapNoCtx 将一个不接收 context 的函数适配为 ExecFunc。
func WrapNoCtx(fn func() error) ExecFunc {
	return func(ctx context.Context) error { return fn() }
}

// WrapNoCtxErr 将一个既不接收 context 也不返回 error 的函数适配为 ExecFunc。
func WrapNoCtxErr(fn func()) ExecFunc {
	return func(ctx context.Context) error { fn(); return nil }
}

// ExecFunc 是生命周期钩子的执行函数签名。
type ExecFunc = func(context.Context) error

// Executor 包装一个生命周期钩子函数。
type Executor struct {
	Exec ExecFunc
}

// Handler 用于在构建期向 Lifecycle 注册钩子。
type Handler func(lc Lifecycle)

// Lifecycle 提供注册各阶段钩子的能力（写入侧）。
type Lifecycle interface {
	AfterStop(f ExecFunc)
	BeforeStop(f ExecFunc)
	AfterStart(f ExecFunc)
	BeforeStart(f ExecFunc)
}

// Getter 提供读取各阶段已注册钩子的能力（读取侧）。
type Getter interface {
	GetAfterStops() []Executor
	GetBeforeStops() []Executor
	GetAfterStarts() []Executor
	GetBeforeStarts() []Executor
}
