package tracing

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/fastrand"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/request_id"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/watcher"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnWatch: func(name string, resp *watcher.Response) {
			resp.OnPut(func() {
				var cfg = GetDefaultCfg()
				xerror.Panic(types.Decode(resp.Value, &cfg))
				xerror.Panic(cfg.Build())
			})
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
				var tracer = opentracing.GlobalTracer()
				if tracer == nil {
					return xerror.Fmt("tracer is nil")
				}

				var (
					err               error
					span              opentracing.Span
					parentSpanContext opentracing.SpanContext
				)

				if !req.Client() {
					// 服务端请求
					// 从请求header中解析链路信息
					parentSpanContext, err = tracer.Extract(opentracing.TextMap, textMapCarrier(req.Header()))
					if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
						// 百分之一的概率
						if fastrand.Probability(0.01) {
							zap.S().Errorf("opentracing: failed parsing trace information: %v", err)
						}
					}
					span = opentracing.StartSpan(req.Endpoint(), ext.RPCServerOption(parentSpanContext))
				} else {
					// 客户端请求
					// 从context中获取span
					span = opentracing.SpanFromContext(ctx)
					if span != nil {
						parentSpanContext = span.Context()
					}

					span = opentracing.StartSpan(req.Endpoint(), opentracing.ChildOf(parentSpanContext), ext.SpanKindRPCClient)
					if err = tracer.Inject(span.Context(), opentracing.TextMap, textMapCarrier(req.Header())); err != nil {
						zap.S().Errorf("opentracing: failed serializing trace information: %v", err)
					}
				}

				var reqId = request_id.GetReqID(ctx)
				span.SetTag(request_id.Name, reqId)

				defer span.Finish()
				err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
				SetIfErr(span, err)
				return err
			}
		},
	})
}
