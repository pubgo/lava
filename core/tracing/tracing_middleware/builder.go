package tracing_middleware

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"

	requestid2 "github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/service"
)

func Middleware(tracer opentracing.Tracer, log *zap.Logger) service.Middleware {
	log = log.Named(logutil.Names(logkey.Module, tracing.Name))
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, resp service.Response) error {
			var (
				err               error
				span              opentracing.Span
				parentSpanContext opentracing.SpanContext
			)

			// 请求trace解析和注入
			if !req.Client() {
				// 服务端请求
				// 从header中解析链路信息
				parentSpanContext, err = tracer.Extract(opentracing.TextMap, &textMapCarrier{req.Header()})
				if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
					logutil.ErrRecord(log, err)
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
				if err = tracer.Inject(span.Context(), opentracing.TextMap, &textMapCarrier{req.Header()}); err != nil {
					logutil.ErrRecord(log, err)
				}
			}

			// request-id绑定
			span.SetTag(requestid2.Name, requestid2.GetFromCtx(ctx))

			tracing.GetFrom(ctx).SetTag("sss", "")

			defer span.Finish()
			err = next(opentracing.ContextWithSpan(ctx, span), req, resp)
			tracing.SetIfErr(span, err)
			return err
		}
	}
}
