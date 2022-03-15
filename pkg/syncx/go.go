package syncx

import (
	"context"
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/logging/logutil"
)

// GoChan 通过chan的方式同步执行异步任务
func GoChan(fn func() Value) chan Value {
	checkFn(fn, "[GoChan] [fn] is nil")

	var ch = make(chan Value)

	go func() {
		defer func() {
			xerror.Resp(func(err xerror.XErr) {
				ch <- WithErr(xerror.Wrap(err, "GoChan", stack.Func(fn)))
			})
			close(ch)
		}()

		if val := fn(); val == nil {
			ch <- nil
		} else {
			ch <- val
		}
	}()

	return ch
}

// GoSafe 安全并发处理
func GoSafe(fn func(), cb ...func(err error)) {
	checkFn(fn, "[GoSafe] [fn] is nil")

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			if len(cb) > 0 {
				logutil.ErrTry(logs.L(), func() { cb[0](err) })
				return
			}

			logErr(fn, err)
		})

		fn()
	}()
}

// GoCtx 可取消并发处理
func GoCtx(fn func(ctx context.Context), cb ...func(err error)) context.CancelFunc {
	checkFn(fn, "[GoCtx] [fn] is nil")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			if len(cb) > 0 {
				logutil.ErrTry(logs.L(), func() { cb[0](err) })
				return
			}

			logErr(fn, err)
		})

		fn(ctx)
	}()

	return cancel
}

// GoDelay 异步延迟处理
func GoDelay(fn func(), durations ...time.Duration) {
	checkFn(fn, "[GoDelay] [fn] is nil")

	dur := time.Millisecond * 10
	if len(durations) > 0 {
		dur = durations[0]
	}

	xerror.Assert(dur == 0, "[dur] should not be 0")

	go func() {
		logutil.ErrTry(logs.L(), fn)
	}()

	time.Sleep(dur)

	return
}

func Monitor(timeout time.Duration, run func(), errFn func(err error)) {
	if timeout <= 0 {
		panic("[Monitor] [timeout] should not be less than zero")
	}

	checkFn(run, "[Monitor] [run] is nil")
	checkFn(errFn, "[Monitor] [errFn] is nil")

	var done = make(chan struct{})
	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			logutil.ErrTry(logs.L(), func() { errFn(err) }, logutil.FuncStack(run))
		})

		run()
		close(done)
	}()

	for {
		select {
		case <-time.After(timeout):
			logutil.ErrTry(logs.L(), func() { errFn(context.DeadlineExceeded) }, logutil.FuncStack(run))
		case <-done:
			return
		}
	}
}

// Timeout 超时处理
func Timeout(dur time.Duration, fn func()) (gErr error) {
	defer xerror.RespErr(&gErr)

	if dur <= 0 {
		panic("[Timeout] [dur] should not be less than zero")
	}

	checkFn(fn, "[Timeout] [fn] is nil")

	var done = make(chan struct{})

	go func() {
		defer func() {
			xerror.RespErr(&gErr)
			close(done)
		}()

		fn()
	}()

	select {
	case <-time.After(dur):
		return context.DeadlineExceeded
	case <-done:
		return
	}
}

func logErr(fn interface{}, err xerror.XErr) {
	logs.WithErr(err).With(logutil.FuncStack(fn)).Error(err.Error())
}

func checkFn(fn interface{}, msg string) {
	if fn == nil {
		panic(msg)
	}
}
