package hook

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/pubgo/lava/plugins/tracing"
)

const MaxPipelineNameCmdCount = 3

func newHook(options *redis.Options) *hook {
	return &hook{
		options: options,
	}
}

type hook struct {
	options *redis.Options
}

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	var span = tracing.CreateChild(ctx, "redis.client")
	return span.WithCtx(ctx), nil
}

func (h *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	var span = tracing.FromCtx(ctx)

	defer func() {
		method, key := h.wrapperName(cmd)
		h.setTag(span, method, key)
		if cmd.Err() != nil {
			tracing.SetIfErr(span, cmd.Err())
		}
		span.Finish()
	}()

	return nil
}

func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	var span = tracing.CreateChild(ctx, "redis.client.pipeline")
	return span.WithCtx(ctx), nil
}

func (h *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := tracing.FromCtx(ctx)

	defer func() {
		method, key := h.wrapperPipelineName(cmds)
		h.setTag(span, method, key)
		// 参考redis驱动本身处理cmds error,只取第一个
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil {
				tracing.SetIfErr(span, cmd.Err())
				break
			}
		}
		span.Finish()
	}()

	return nil
}

func (h *hook) setTag(span opentracing.Span, method, key string) {
	ext.DBType.Set(span, "redis")
	if h.options != nil {
		ext.PeerAddress.Set(span, h.options.Addr)
	}
	ext.SpanKind.Set(span, "redis-client")

	// add redis command
	span.SetTag("db.method", method)
	span.SetTag("db.key", key)
}

func (h *hook) wrapperName(cmd redis.Cmder) (name string, key string) {
	name = cmd.Name()
	args := cmd.Args()
	if len(args) > 1 {
		k, ok := args[1].(string)
		if ok {
			key = k
		}
	}
	return
}

func (h *hook) wrapperPipelineName(cmders []redis.Cmder) (name string, key string) {
	names := make([]string, len(cmders))
	keys := make([]string, len(cmders))
	for i, cmd := range cmders {
		// keep MaxPipelineNameCmdCount cmd name
		if i >= MaxPipelineNameCmdCount {
			break
		}
		names[i], keys[i] = h.wrapperName(cmd)
	}

	name = strings.Join(names, "->")
	key = strings.Join(keys, "->")

	return
}
