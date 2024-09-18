package cloudjobs

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/stack"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/lava/internal/ctxutil"
	"github.com/pubgo/lava/pkg/proto/cloudjobpb"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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

func (c *Client) Publish(ctx context.Context, key string, args proto.Message, opts ...*cloudjobpb.PushEventOptions) error {
	return c.publish(ctx, key, args, opts...)
}

func (c *Client) publish(ctx context.Context, key string, args proto.Message, opts ...*cloudjobpb.PushEventOptions) (gErr error) {
	var timeout = ctxutil.GetTimeout(ctx)
	var now = time.Now()
	var msgId = xid.New().String()
	var pushEventOpt *cloudjobpb.PushEventOptions

	defer func() {
		var msgFn = func(e *zerolog.Event) {
			e.Str("pub_topic", key)
			e.Str("pub_start", now.String())
			e.Any("pub_args", args)
			e.Str("pub_cost", time.Since(now).String())
			e.Str("pub_msg_id", msgId)
			if timeout != nil {
				e.Str("timeout", timeout.String())
			}
		}
		if gErr == nil {
			logger.Info(ctx).Func(msgFn).Msg("succeed to publish cloud job event to stream")
		} else {
			logger.Err(gErr, ctx).Func(msgFn).Msg("failed to publish cloud job event to stream")
		}
	}()

	pushEventOpt = getOptions(ctx, opts...)

	if pushEventOpt.MsgId != nil {
		msgId = pushEventOpt.GetMsgId()
	}

	pb, err := anypb.New(args)
	if err != nil {
		return errors.Wrap(err, "failed to marshal args to any proto")
	}

	// TODO get info from ctx
	data, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrap(err, "failed to marshal any proto to bytes")
	}

	// subject|topic name
	key = c.subjectName(key)

	msg := &nats.Msg{
		Subject: key,
		Data:    data,
		Header: nats.Header{
			senderKey:        []string{fmt.Sprintf("%s/%s", running.Project, running.Version)},
			cloudJobDelayKey: []string{encodeDelayTime(pushEventOpt.DelayDur.AsDuration())},
		},
	}

	for k, v := range pushEventOpt.Metadata {
		msg.Header.Add(k, v)
	}

	jetOpts := append([]jetstream.PublishOpt{}, jetstream.WithMsgID(msgId))
	_, err = c.js.PublishMsg(ctx, msg, jetOpts...)
	if err != nil {
		return errors.Wrapf(err, "failed to publish msg to stream, topic=%s msg_id=%s", key, msgId)
	}

	return nil
}
