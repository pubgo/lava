package lifecyclebuilder

import (
	"context"
	"testing"

	"github.com/pubgo/lava/v2/core/lifecycle"
)

func execAll(t *testing.T, execs []lifecycle.Executor) {
	t.Helper()
	for _, e := range execs {
		if err := e.Exec(context.Background()); err != nil {
			t.Fatalf("exec returned error: %v", err)
		}
	}
}

// TestStartHooksFIFO 验证启动钩子按注册顺序（FIFO）执行。
func TestStartHooksFIFO(t *testing.T) {
	var order []int
	p := New([]lifecycle.Handler{
		func(lc lifecycle.Lifecycle) {
			lc.BeforeStart(lifecycle.WrapNoCtxErr(func() { order = append(order, 1) }))
			lc.BeforeStart(lifecycle.WrapNoCtxErr(func() { order = append(order, 2) }))
			lc.BeforeStart(lifecycle.WrapNoCtxErr(func() { order = append(order, 3) }))
		},
	})

	execAll(t, p.Getter.GetBeforeStarts())

	want := []int{1, 2, 3}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("start hooks should be FIFO, got %v want %v", order, want)
		}
	}
}

// TestStopHooksLIFO 验证停止钩子按注册逆序（LIFO）执行。
func TestStopHooksLIFO(t *testing.T) {
	var order []int
	p := New([]lifecycle.Handler{
		func(lc lifecycle.Lifecycle) {
			lc.AfterStop(lifecycle.WrapNoCtxErr(func() { order = append(order, 1) }))
			lc.AfterStop(lifecycle.WrapNoCtxErr(func() { order = append(order, 2) }))
			lc.AfterStop(lifecycle.WrapNoCtxErr(func() { order = append(order, 3) }))
		},
	})

	execAll(t, p.Getter.GetAfterStops())

	want := []int{3, 2, 1}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("stop hooks should be LIFO, got %v want %v", order, want)
		}
	}
}
