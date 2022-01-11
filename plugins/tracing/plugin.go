package tracing

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugin"
	requestID2 "github.com/pubgo/lava/plugins/requestID"
	"github.com/pubgo/lava/types"
)

var logs = logz.Component(Name)

const Name = "tracing"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
				var tracer = opentracing.GlobalTracer()
				if tracer == nil {
					logs.Warn("global tracer is nil, please init tracing")
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
				span.SetTag(requestID2.Name, requestID2.GetWith(ctx))

				defer span.Finish()
				err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
				SetIfErr(span, err)
				return err
			}
		},
		OnVars: func(v types.Vars) {
			v.Do(Name+"_cfg", func() interface{} { return cfg })
			v.Do(Name+"_factory", func() interface{} {
				var data = make(map[string]string)
				factories.Range(func(key, value interface{}) bool {
					data[key.(string)] = stack.Func(value)
					return true
				})
				return data
			})
		},
	})
}
