package requestid

import (
	"context"
	"github.com/pubgo/lava/service"

	"github.com/pubgo/dix"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/internal/pkg/httpx"
	"github.com/pubgo/lava/internal/pkg/utils"
)

const Name = "x-request-id"

func init() {
	dix.Register(func() service.Middlewares {
		return service.Middlewares{
			func(next service.HandlerFunc) service.HandlerFunc {
				return func(ctx context.Context, req service.Request, resp service.Response) (gErr error) {
					defer func() {
						switch err := recover().(type) {
						case nil:
						case error:
							gErr = err
						default:
							gErr = status.Errorf(codes.Internal, "service=>%s, endpoint=>%s, msg=>%v", req.Service(), req.Endpoint(), err)
						}
					}()

					rid := utils.FirstFnNotEmpty(
						func() string { return getReqID(ctx) },
						func() string { return string(req.Header().Peek(Name)) },
						func() string { return ksuid.New().String() },
					)

					req.Header().Set(httpx.HeaderXRequestID, rid)
					return next(WithReqID(ctx, rid), req, resp)
				}
			},
		}
	})
}
