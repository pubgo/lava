package tracing_plugin

import (
	"context"
	"errors"
	"github.com/pubgo/lava/core/watcher"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/requestID"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/vars"
)

var logs = logging.Component(tracing.Name)

func init() {
	var cfg = tracing.DefaultCfg()
	plugin.Register(&plugin.Base{
		Name: tracing.Name,
		OnInit: func(p plugin.Process) {
			_ = config.Decode(tracing.Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnWatch: func(_ string, r *watcher.Response) error {
			_ = config.Decode(tracing.Name, &cfg)
			return cfg.Build()
		},
		OnMiddleware: func(next service.HandlerFunc) service.HandlerFunc {
			return func(ctx context.Context, req service.Request, resp func(rsp service.Response) error) error {
				var tracer = opentracing.GlobalTracer()
				if tracer == nil {
					logs.L().Warn("global tracer is nil, please init tracing")
					return nil
				}

				var (
					err               error
					span              opentracing.Span
					parentSpanContext opentracing.SpanContext
				)

				// 请求trace解析和注入
				if !req.Client() {
					// 服务端请求
					// 从header中解析链路信息
					parentSpanContext, err = tracer.Extract(opentracing.TextMap, textMapCarrier(req.Header()))
					if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
						logs.WithErr(err).Error("opentracing: failed parsing trace information")
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
						logs.WithErr(err).Error("opentracing: failed serializing trace information")
					}
				}

				// request-id绑定
				span.SetTag(requestID.Name, requestID.GetWith(ctx))

				tracing.GetFrom(ctx).SetTag("sss", "")

				defer span.Finish()
				err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
				tracing.SetIfErr(span, err)
				return err
			}
		},
		OnVars: func(v vars.Publisher) {
			v.Do(tracing.Name+"_cfg", func() interface{} { return cfg })
		},
	})
}
