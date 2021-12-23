package traceRecord

import (
	"context"
	"errors"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/types"
)

const Name = "traceRecord"

var logs = logz.Component(Name)

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
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
			var reqId = requestID.GetWith(ctx)
			span.SetTag(requestID.Name, reqId)

			defer span.Finish()
			err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
			tracing.SetIfErr(span, err)

			return err
		}
	})
}
