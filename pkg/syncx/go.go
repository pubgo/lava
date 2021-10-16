package syncx

import (
	"context"
	"errors"
	"runtime"
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/pkg/fastrand"
)

func checkConcurrent(name string, fn interface{}) func() {
	curConcurrent.Inc()

	// 阻塞, 等待任务处理完毕
	for curConcurrent.Load() > maxConcurrent {
		runtime.Gosched()

		// 百分之一的采样, 打印log, 让监控获取信息
		if fastrand.Sampling(0.01) {
			logger.GetName(Name).Error(
				"The current concurrent number exceeds the maximum concurrent number of the system",
				zap.String("name", name),
				zap.Uint32("current", curConcurrent.Load()),
				zap.Uint32("max", maxConcurrent),
				zap.String("fn", stack.Func(fn)),
			)
		}
	}

	return func() { curConcurrent.Dec() }
}

func logErr(name string, fn interface{}, err xerror.XErr) {
	logger.GetName(Name).Error(
		err.Error(),
		zap.String("name", name),
		zap.String("fn", stack.Func(fn)),
		zap.String("err_stack", err.Error()),
	)
}

// GoChan 通过chan的方式同步执行并发任务
func GoChan(fn func(), cb ...func(err error)) chan struct{} {
	if fn == nil {
		panic("[GoChan] [fn] is nil")
	}

	var dec = checkConcurrent("GoChan", fn)

	var ch = make(chan struct{})

	go func() {
		close(ch)
		defer xerror.Resp(func(err xerror.XErr) {
			// 过滤Canceled类型的错误
			if errors.Is(err, context.Canceled) {
				return
			}

			logErr("GoChan", fn, err)

			if len(cb) > 0 {
				defer xerror.RespExit()
				cb[0](err)
			}
		})
		defer dec()

		fn()
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

			logErr("GoSafe", fn, err)

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

			logErr("GoCtx", fn, err)

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

	GoSafe(fn, func(err error) { dur = 0 })

	if dur != 0 {
		time.Sleep(dur)
	}

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

			logErr("GoTimeout", fn, err)
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

func newPromise() *Promise { return &Promise{ch: make(chan interface{})} }

type Promise struct {
	err error
	ch  chan interface{}
}

func (t *Promise) Err() error                  { return t.err }
func (t *Promise) close()                      { close(t.ch) }
func (t *Promise) Unwrap() <-chan interface{}  { return t.ch }
func (t *Promise) Await() (interface{}, error) { return <-t.ch, t.err }

func Yield(fn func() (interface{}, error)) *Promise {
	if fn == nil {
		panic("[Yield] [fn] should not be nil")
	}

	var p = newPromise()
	GoSafe(func() {
		defer p.close()
		defer xerror.RespErr(&p.err)
		val, err := fn()
		p.err = err
		p.ch <- val
	})

	return p
}

func YieldMap(fn func(in chan<- *Promise) error) *Promise {
	if fn == nil {
		panic("[YieldMap] [fn] should not be nil")
	}

	var p = &Promise{ch: make(chan interface{})}
	var in = make(chan *Promise)

	GoSafe(func() {
		defer p.close()
		for pp := range in {
			for val := range pp.Unwrap() {
				p.ch <- val
			}

			if pp.Err() != nil {
				p.err = pp.Err()
			}
		}
	})

	GoSafe(func() {
		defer close(in)
		defer xerror.RespErr(&p.err)
		p.err = fn(in)
	})

	return p
}
