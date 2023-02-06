package requestid

import (
	"context"

	"github.com/pubgo/funk/strutil"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/pkg/httputil"
	"github.com/pubgo/lava/service"
)

const Name = "x-request-id"

func Middleware() service.Middleware {
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, rsp service.Response) (gErr error) {
			defer func() {
				switch err := recover().(type) {
				case nil:
				case error:
					gErr = err
				default:
					gErr = status.Errorf(codes.Internal, "service=>%s, endpoint=>%s, msg=>%v", req.Service(), req.Endpoint(), err)
				}
			}()

			rid := strutil.FirstFnNotEmpty(
				func() string { return getReqID(ctx) },
				func() string { return string(req.Header().Peek(Name)) },
				func() string { return xid.New().String() },
			)

			req.Header().Set(httputil.HeaderXRequestID, rid)
			return next(CreateCtx(ctx, rid), req, rsp)
		}
	}
}
