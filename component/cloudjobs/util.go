package eventjob

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/lava/internal/ctxutil"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func RegisterJobHandler[T proto.Message](jobCli *Client, jobName string, topic string, handler JobHandler[T]) {
	assert.Fn(reflect.TypeOf(subjects[topic]) != reflect.TypeOf(lo.Empty[T]()), func() error {
		return fmt.Errorf("type not match, topic-type=%s handler-input-type=%s", reflect.TypeOf(subjects[topic]).String(), reflect.TypeOf(lo.Empty[T]()).String())
	})

	jobCli.registerJobHandler(jobName, topic, func(ctx *Context, args proto.Message) error { return handler(ctx, args.(T)) })
}

// PushEventSync push event sync
func PushEventSync[T any](handler func(context.Context, T) (*emptypb.Empty, error), ctx context.Context, t T) error {
	return pushEventBasic(handler, ctx, t, false)
}

// PushEvent push event async
func PushEvent[T any](handler func(context.Context, T) (*emptypb.Empty, error), ctx context.Context, t T, errs ...chan error) {
	// clone ctx and recalculate timeout
	ctx, cancel := ctxutil.Clone(ctx, DefaultTimeout)
	defer cancel()

	err := pushEventBasic(handler, ctx, t, true)
	if err == nil || len(errs) == 0 {
		return
	}

	select {
	case errs[0] <- err:
		return
	default:
		log.Warn(ctx).Func(func(e *zerolog.Event) {
			e.Str("fn_caller", stack.CallerWithFunc(handler).String())
			e.Msg("failed to receive error message with push event")
		})
	}
}

func pushEventBasic[T any](handler func(context.Context, T) (*emptypb.Empty, error), ctx context.Context, t T, async bool) error {
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

		e.Str("fn_caller", stack.Caller(1).String())
		e.Str("deal", stack.Caller(1).String())
		e.Bool("async", async)
		e.Any("input_params", t)
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
