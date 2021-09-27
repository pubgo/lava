package tracing

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/watcher"
)

func init() {
	plugin.Register(&plugin.Base{
		Name:         Name,
		OnMiddleware: Middleware,
		OnInit: func(ent plugin.Entry) {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnWatch: func(name string, resp *watcher.Response) {
			resp.OnPut(func() {
				var cfg = GetDefaultCfg()
				xerror.Panic(watcher.Decode(resp.Value, &cfg))
				xerror.Panic(cfg.Build())
			})
		},
	})
}

func Middleware(next types.MiddleNext) types.MiddleNext {
	return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
		var tracer = opentracing.GlobalTracer()
		if tracer == nil {
			return xerror.Fmt("tracer is nil")
		}

		var (
			span              opentracing.Span
			err               error
			parentSpanContext opentracing.SpanContext
		)

		if !req.Client() {
			// 服务端tracing, 从header中解析链路信息
			parentSpanContext, err = tracer.Extract(opentracing.TextMap, textMapCarrier(req.Header()))
			if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
				zap.S().Errorf("opentracing: failed parsing trace information: %v", err)
			}
			span = opentracing.StartSpan(req.Endpoint(), ext.RPCServerOption(parentSpanContext))
		} else {
			// 客户端tracing, 从context获取span
			span = opentracing.SpanFromContext(ctx)
			if span != nil {
				parentSpanContext = span.Context()
			}

			span = opentracing.StartSpan(req.Endpoint(), opentracing.ChildOf(parentSpanContext), ext.SpanKindRPCClient)
			if err = tracer.Inject(span.Context(), opentracing.TextMap, textMapCarrier(req.Header())); err != nil {
				zap.S().Errorf("opentracing: failed serializing trace information: %v", err)
			}
		}

		return next(opentracing.ContextWithSpan(ctx, span), req, resp)
	}
}
