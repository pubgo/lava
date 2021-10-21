package traceRecord

import (
	"context"
	"errors"
	"github.com/pubgo/lava/plugins/tracing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/pkg/fastrand"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Middleware("traceRecord", func(next types.MiddleNext) types.MiddleNext {
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
					if fastrand.Probability(10) {
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

			var reqId = requestID.GetWith(ctx)
			span.SetTag(requestID.Name, reqId)

			defer span.Finish()
			err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
			tracing.SetIfErr(span, err)
			return err
		}
	})
}
