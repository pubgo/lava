package cloudjobs

import (
	"context"
	"time"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/lava/internal/ctxutil"
	"github.com/pubgo/lava/pkg/proto/cloudjobpb"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"
)

func PushEventWithOpt[T any](handler func(*Client, context.Context, T, ...*cloudjobpb.PushEventOptions) error, jobCli *Client, ctx context.Context, t T, opts ...*cloudjobpb.PushEventOptions) chan error {
	errChan := make(chan error, 1)
	timeout := ctxutil.GetTimeout(ctx)
	now := time.Now()
	fnCaller := stack.Caller(1).String()

	// clone ctx and recalculate timeout
	ctx = lo.T2(ctxutil.Clone(ctx, DefaultTimeout)).A
	doHandler := func() error {
		err := try.Try(func() error { return errors.WrapCaller(handler(jobCli, ctx, t, opts...)) })
		if err == nil {
			return nil
		}

		logger.Err(err, ctx).Func(func(e *zerolog.Event) {
			if timeout != nil {
				e.Str("timeout", timeout.String())
			}

			e.Str("fn_caller", fnCaller)
			e.Bool("async", true)
			e.Any("input", t)
			e.Str("stack", stack.CallerWithFunc(handler).String())
			e.Str("cost", time.Since(now).String())
			e.Msg("failed to push event msg to nats stream")
		})
		return err
	}
	go func() { errChan <- doHandler() }()
	return errChan
}

// PushEvent push event async
func PushEvent[T any](handler func(context.Context, T) (*emptypb.Empty, error), ctx context.Context, t T, opts ...*cloudjobpb.PushEventOptions) chan error {
	// clone ctx and recalculate timeout
	ctx = lo.T2(ctxutil.Clone(ctx, DefaultTimeout)).A
	fnCaller := stack.Caller(1).String()
	errChan := make(chan error, 1)

	if len(opts) > 0 {
		ctx = withOptions(ctx, opts[0])
	}

	go func() { errChan <- pushEventBasic(handler, ctx, t, true, fnCaller) }()
	return errChan
}

func pushEventBasic[T any](handler func(context.Context, T) (*emptypb.Empty, error), ctx context.Context, t T, async bool, caller string) error {
	timeout := ctxutil.GetTimeout(ctx)
	now := time.Now()
	err := try.Try(func() error { return lo.T2(handler(ctx, t)).B })
	if err == nil {
		return nil
	}

	logger.Err(err, ctx).Func(func(e *zerolog.Event) {
		if timeout != nil {
			e.Str("timeout", timeout.String())
		}

		e.Str("fn_caller", caller)
		e.Bool("async", async)
		e.Any("input", t)
		e.Str("stack", stack.CallerWithFunc(handler).String())
		e.Str("cost", time.Since(now).String())
		e.Msg("failed to push event msg to nats stream")
	})
	return err
}
