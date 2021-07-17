package tracing

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/x/stack"

	"github.com/pubgo/lug/entry"
)

func Middleware() entry.Middleware {
	return func(next entry.Wrapper) entry.Wrapper {
		return func(ctx context.Context, req entry.Request, resp func(rsp interface{})) error {
			var span = opentracing.SpanFromContext(ctx)
			if span == nil {
				span = opentracing.StartSpan(req.Endpoint())
			}

			fmt.Println(stack.Func(next))
			return next(withCtx(ctx, NewSpan(span)), req, resp)
		}
	}
}
