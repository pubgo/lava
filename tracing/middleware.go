package tracing

import (
	"context"
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/lug/types"
)

func Middleware() types.Middleware {
	return func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
			var tracer = opentracing.GlobalTracer()
			if tracer == nil {
				return xerror.Fmt("tracer is nil")
			}

			var span *Span

			if !req.Client() {
				parentSpanContext, err := tracer.Extract(opentracing.TextMap, textMapCarrier(req.Header()))
				if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
					grpclog.Infof("opentracing: failed parsing trace information: %v", err)
				}
				span = StartSpan(req.Endpoint(), ext.RPCServerOption(parentSpanContext))
			} else {
				span = FromCtx(ctx)
				var parentSpanCtx opentracing.SpanContext
				if span != nil {
					parentSpanCtx = span.Context()
				}

				span = StartSpan(req.Endpoint(), opentracing.ChildOf(parentSpanCtx), ext.SpanKindRPCClient)
				if err := tracer.Inject(span.Context(), opentracing.TextMap, textMapCarrier(req.Header())); err != nil {
					grpclog.Infof("opentracing: failed serializing trace information: %v", err)
				}
			}

			defer func() {
				SetIfErr(span, gErr)

				span.Finish()
			}()

			gErr = next(withCtx(ctx, span), req, resp)
			return
		}
	}
}
