package middleware_recovery

import (
	"context"

	"github.com/pubgo/funk/v2/errors"

	"github.com/pubgo/lava/v2/lava"
)

func New() lava.Middleware {
	return lava.MiddlewareWrap{
		Name: "recovery",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
				defer func() {
					gErr = errors.WrapStack(errors.Parse(recover()))
				}()

				return next(ctx, req)
			}
		},
	}
}
