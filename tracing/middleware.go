package tracing

import (
	"context"
	"errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/lug/types"
)

func Middleware() types.Middleware {
	return func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			var tracer = opentracing.GlobalTracer()
			if tracer == nil {
				return xerror.Fmt("tracer is nil")
			}

			var (
				span              *Span
				err               error
				parentSpanContext opentracing.SpanContext
			)

			if !req.Client() {
				parentSpanContext, err = tracer.Extract(opentracing.TextMap, textMapCarrier(req.Header()))
				if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
					grpclog.Infof("opentracing: failed parsing trace information: %v", err)
				}

				span = StartSpan(req.Endpoint(), ext.RPCServerOption(parentSpanContext))
			} else {
				span = FromCtx(ctx)
				if !span.Noop() {
					parentSpanContext = span.Context()
				}

				span = StartSpan(req.Endpoint(), opentracing.ChildOf(parentSpanContext), ext.SpanKindRPCClient)
				if err := tracer.Inject(span.Context(), opentracing.TextMap, textMapCarrier(req.Header())); err != nil {
					grpclog.Infof("opentracing: failed serializing trace information: %v", err)
				}
			}

			ctx = opentracing.ContextWithSpan(ctx, span)
			ctx = xlog.AppendCtx(ctx, zap.String(TraceId, span.TraceID()))

			defer func() {
				SetIfErr(span, err)
				span.Finish()
			}()

			err = next(ctx, req, resp)
			return err
		}
	}
}
