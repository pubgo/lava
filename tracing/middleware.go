package tracing

import (
	"context"
	"github.com/pubgo/lug/types"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lug/entry"
)

func Middleware() entry.Middleware {
	return func(next entry.Wrapper) entry.Wrapper {
		return func(ctx context.Context, req types.Request, resp func(rsp interface{}) error) error {
			var span = opentracing.SpanFromContext(ctx)
			if span == nil {
				span = opentracing.StartSpan(req.Endpoint())
			}

			return next(withCtx(ctx, NewSpan(span)), req, resp)
		}
	}
}
