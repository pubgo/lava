package middleware_service

import (
	"context"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/version"
	lavapbv1 "github.com/pubgo/lava/pkg/proto/lava"

	"google.golang.org/grpc/metadata"

	"github.com/pubgo/lava"
)

func New() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			if req.Client() {
				defer func() {
					info := lava.GetServerInfo(ctx)
					metadata.AppendToOutgoingContext(ctx, "", "")
				}()
			} else {
				clientInfo := new(lavapbv1.ServiceInfo)
				metadata.FromIncomingContext(ctx)

				ctx = lava.CreateCtxWithServerInfo(ctx, &lavapbv1.ServiceInfo{
					Name:      version.Project(),
					Version:   version.Version(),
					Path:      req.Operation(),
					Hostname:  runmode.Hostname,
					RequestId: lava.GetReqID(ctx),
				})
			}

			return next(ctx, req)
		}
	}
}
