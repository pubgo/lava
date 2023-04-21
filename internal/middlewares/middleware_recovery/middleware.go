package middleware_recovery

import (
	"context"
	"runtime/debug"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/proto/errorpb"

	"github.com/pubgo/lava"
)

func New() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			defer func() {
				if err := errors.Parse(recover()); err != nil {
					debug.PrintStack()
					gErr = errors.WrapStack(err)
				}
			}()

			if v, ok := req.Payload().(lava.Validator); ok && v != nil {
				if e := log.Ctx(ctx).Debug(); e.Enabled() {
					e.Msg("validate request")
				}

				gErr = v.Validate()
				if gErr != nil {
					return nil, errors.WrapCode(gErr, &errorpb.ErrCode{
						Code:   errorpb.Code_InvalidArgument,
						Status: "lava.request.validate",
						Reason: gErr.Error(),
					})
				}
			}

			return next(ctx, req)
		}
	}
}
