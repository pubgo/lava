package middleware_service

import (
	"context"
	"github.com/pubgo/funk/runmode"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/grpcutil"
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
					req.Header().Set(grpcutil.ClientHostnameKey, info.Name)
					req.Header().Set(grpcutil.ClientHostnameKey, info.Version)
					req.Header().Set(grpcutil.ClientHostnameKey, info.Path)
					req.Header().Set(grpcutil.ClientHostnameKey, info.Hostname)
					req.Header().Set(grpcutil.ClientHostnameKey, info.Ip)
				}()
			} else {
				clientInfo := new(lavapbv1.ServiceInfo)
				metadata.FromIncomingContext(ctx)

				ctx = lava.CreateCtxWithServerInfo(ctx, &lavapbv1.ServiceInfo{
					Name:     version.Project(),
					Version:  version.Version(),
					Path:     req.Operation(),
					Hostname: runmode.Hostname,
				})
			}

			return next(ctx, req)
		}
	}
}
