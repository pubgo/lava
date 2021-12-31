package syncx

import (
	"context"
	"errors"
	"github.com/pubgo/lava/logger/logutil"
	"runtime"
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/pkg/fastrand"
)

// GoChan 通过chan的方式同步执行异步任务
func GoChan(fn func() Value) chan Value {
	if fn == nil {
		panic("[GoChan] [fn] is nil")
	}

	var dec = checkConcurrent("GoChan", fn)

	var ch = make(chan Value)

	go func() {
		defer func() {
			xerror.Resp(func(err xerror.XErr) {
				// 过滤Canceled类型的错误
				if errors.Is(err, context.Canceled) {
					return
				}

				ch <- WithErr(xerror.Wrap(err, "GoChan", stack.Func(fn)))
			})
			close(ch)
			dec()
		}()

		var val = fn()
		if val == nil {
			ch <- &valueImpl{}
		} else {
			ch <- val
		}
	}()

	return ch
}

// GoErr 异常处理安全并发
func GoErr(err *error, fn func()) {
	if fn == nil {
		panic("[GoSafe] [fn] is nil")
	}

	var dec = checkConcurrent("GoErr", fn)

	go func() {
		defer xerror.RespErr(err)
		defer dec()
		fn()
	}()
}

// GoSafe 安全并发处理
func GoSafe(fn func(), cb ...func(err error)) {
	if fn == nil {
		panic("[GoSafe] [fn] is nil")
	}

	var dec = checkConcurrent("GoSafe", fn)

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			// 过滤Canceled类型的错误
			if errors.Is(err, context.Canceled) {
				return
			}

			logErr(fn, err)

			if len(cb) > 0 {
				defer xerror.RespExit()
				cb[0](err)
			}
		})
		defer dec()

		fn()
	}()
}

// GoCtx 可取消并发处理
func GoCtx(fn func(ctx context.Context), cb ...func(err error)) context.CancelFunc {
	if fn == nil {
		panic("[GoCtx] [fn] is nil")
	}

	var dec = checkConcurrent("GoCtx", fn)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			// 过滤Canceled类型的错误
			if errors.Is(err, context.Canceled) {
				return
			}

			logErr(fn, err)

			if len(cb) > 0 {
				defer xerror.RespExit()
				cb[0](err)
			}
		})
		defer dec()

		fn(ctx)
	}()

	return cancel
}

// GoDelay 延迟并发处理
func GoDelay(fn func(), durations ...time.Duration) {
	dur := time.Millisecond * 10
	if len(durations) > 0 {
		dur = durations[0]
	}

	xerror.Assert(dur == 0, "[dur] should not be 0")

	time.Sleep(dur)

	GoSafe(fn)

	return
}

// GoTimeout 超时处理
func GoTimeout(dur time.Duration, fn func()) (gErr error) {
	defer xerror.RespErr(&gErr)

	if dur <= 0 {
		panic("[GoTimeout] [dur] should not be less than zero")
	}

	if fn == nil {
		panic("[GoTimeout] [fn] is nil")
	}

	var dec = checkConcurrent("GoTimeout", fn)

	var ch = make(chan struct{})

	go func() {
		close(ch)
		defer xerror.Resp(func(err xerror.XErr) {
			// 过滤Canceled类型的错误
			if errors.Is(err, context.Canceled) {
				return
			}

			gErr = err

			logErr(fn, err)
		})
		defer dec()

		fn()
	}()

	select {
	case <-ch:
		return
	case <-time.After(dur):
		return context.DeadlineExceeded
	}
}

// checkConcurrent 检查当前goroutine数量
func checkConcurrent(name string, fn interface{}) func() {
	curConcurrent.Inc()

	// 阻塞,等待任务执行完毕
	for curConcurrent.Load() > maxConcurrent {
		runtime.Gosched()

		// 采样率(1%), 打印log, 让监控获取信息
		if fastrand.Sampling(0.01) {
			logs.With(
				zap.String("name", name),
				zap.Int64("current", curConcurrent.Load()),
				zap.Int64("maximum", maxConcurrent),
				logutil.FuncStack(fn),
			).Error("The current concurrent number exceeds the maximum concurrent number of the system")
		}
	}

	return func() { curConcurrent.Dec() }
}

func logErr(fn interface{}, err xerror.XErr) {
	logs.WithErr(err, logutil.FuncStack(fn)).Error(err.Error())
}
