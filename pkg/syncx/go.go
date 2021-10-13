package syncx

import (
	"context"
	"errors"
	"runtime"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/pkg/fastrand"
)

func SafeGo(fn func(), cb ...func(err error)) {
	if fn == nil {
		panic("[fn] is nil")
	}

	curConcurrent.Inc()

	// 阻塞, 等待任务处理完毕
	for curConcurrent.Load() > maxConcurrent {
		runtime.Gosched()

		// 百分之一的概率, 打印log
		if fastrand.Probability(10) {
			logger.GetName(Name).Sugar().Errorf("The current(%d) concurrent number exceeds the maximum(%d) concurrent number of the system", curConcurrent.Load(), maxConcurrent)
		}
	}

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			// Canceled 错误类型过滤
			if errors.Is(err, context.Canceled) {
				return
			}

			logger.GetName(Name).Error("[SafeGo] handler error", zap.Any("err", err), zap.String("err_msg", err.Error()))

			if len(cb) > 0 {
				xerror.Exit(xerror.Try(func() { cb[0](err) }))
			}
		})
		defer curConcurrent.Dec()

		fn()
	}()
}

func CtxGo(fn func(ctx context.Context), cb ...func(err error)) context.CancelFunc {
	if fn == nil {
		panic("[fn] is nil")
	}

	curConcurrent.Inc()

	// 阻塞, 等待任务处理完毕
	for curConcurrent.Load() > maxConcurrent {
		runtime.Gosched()

		// 百分之一的概率, 打印log
		if fastrand.Probability(10) {
			logger.GetName(Name).Sugar().Errorf("The current(%d) concurrent number exceeds the maximum(%d) concurrent number of the system", curConcurrent.Load(), maxConcurrent)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer xerror.Resp(func(err xerror.XErr) {
			// Canceled 错误类型过滤
			if errors.Is(err, context.Canceled) {
				return
			}

			logger.GetName(Name).Error("[CtxGo] handler error", zap.Any("err", err), zap.String("err_msg", err.Error()))

			if len(cb) > 0 {
				xerror.Exit(xerror.Try(func() { cb[0](err) }))
			}
		})
		defer curConcurrent.Dec()

		fn(ctx)
	}()

	return cancel
}
