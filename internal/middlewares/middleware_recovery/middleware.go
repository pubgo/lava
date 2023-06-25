package middleware_recovery

import (
	"context"
	"runtime/debug"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/lava"
)

func New() lava.Middleware {
	return lava.MiddlewareWrap{
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
				defer func() {
					if err := errors.Parse(recover()); err != nil {
						debug.PrintStack()
						gErr = errors.WrapStack(err)
					}
				}()

				return next(ctx, req)
			}
		},
		Name: "recovery",
	}
}
