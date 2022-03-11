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

// GoMonitor check timeout
func GoMonitor(timeout time.Duration, ok func() bool, errFn func(err error)) chan struct{} {
	if timeout <= 0 {
		panic("[GoMonitor] [timeout] should not be less than zero")
	}

	checkFn(ok, "[GoMonitor] [fn] is nil")
	checkFn(errFn, "[GoMonitor] [fn] is nil")

	var fn = func() error {
		var done = make(chan struct{})
		var gErr error

		go func() {
			defer func() {
				xerror.RespErr(&gErr)
				close(done)
			}()

			for ok() {
				time.Sleep(time.Millisecond * 10)
			}
		}()

		select {
		case <-time.After(timeout):
			return context.DeadlineExceeded
		case <-done:
			return gErr
		}
	}

	var done = make(chan struct{})
	go func() {
		defer close(done)
		for {
			var err = fn()
			if err == nil {
				return
			}

			logutil.ErrTry(logs.L(), func() { errFn(err) })
		}
	}()
	return done
}

// GoTimeout 超时处理
func GoTimeout(dur time.Duration, fn func()) (gErr error) {
	defer xerror.RespErr(&gErr)

	if dur <= 0 {
		panic("[GoTimeout] [dur] should not be less than zero")
	}

	checkFn(fn, "[GoTimeout] [fn] is nil")

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
