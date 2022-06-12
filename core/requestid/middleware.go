package requestid

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	middleware2 "github.com/pubgo/lava/core/middleware"
	"github.com/pubgo/lava/internal/pkg/httpx"
	"github.com/pubgo/lava/internal/pkg/utils"
)

const Name = "x-request-id"

func init() {
	dix.Register(func() middleware2.Middlewares {
		return middleware2.Middlewares{
			func(next middleware2.HandlerFunc) middleware2.HandlerFunc {
				return func(ctx context.Context, req middleware2.Request, resp middleware2.Response) (gErr error) {
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
