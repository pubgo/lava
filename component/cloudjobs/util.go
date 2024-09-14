package cloudjobs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
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

func getStorageType(name string) jetstream.StorageType {
	switch name {
	case "memory":
		return jetstream.MemoryStorage
	case "file":
		return jetstream.FileStorage
	default:
		panic("unknown storage type")
	}
}

func mergeJobConfig(dst *JobConfig, src *JobConfig) *JobConfig {
	if src == nil {
		src = handleDefaultJobConfig(nil)
	}

	if dst == nil {
		dst = handleDefaultJobConfig(nil)
	}

	if dst.MaxRetry == nil {
		dst.MaxRetry = src.MaxRetry
	}

	if dst.Timeout == nil {
		dst.Timeout = src.Timeout
	}

	if dst.RetryBackoff == nil {
		dst.RetryBackoff = src.RetryBackoff
	}

	return dst
}

func handleDefaultJobConfig(cfg *JobConfig) *JobConfig {
	if cfg == nil {
		cfg = new(JobConfig)
	}

	if cfg.Timeout == nil {
		cfg.Timeout = lo.ToPtr(DefaultTimeout)
	}

	if cfg.MaxRetry == nil {
		cfg.MaxRetry = lo.ToPtr(DefaultMaxRetry)
	}

	if cfg.RetryBackoff == nil {
		cfg.RetryBackoff = lo.ToPtr(DefaultRetryBackoff)
	}

	return cfg
}

func handleSubjectName(name string, prefix string) string {
	prefix = fmt.Sprintf("%s.", prefix)
	if strings.HasPrefix(name, prefix) {
		return name
	}

	return fmt.Sprintf("%s%s", prefix, name)
}

func encodeDelayTime(duration time.Duration) string {
	return strconv.Itoa(int(time.Now().Add(duration).UnixMilli()))
}

func decodeDelayTime(delayTime string) (time.Duration, error) {
	tt, err := strconv.Atoi(delayTime)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse cloud job delay time, time=%s", delayTime)
	}
	return time.Until(time.UnixMilli(int64(tt))), nil
}
