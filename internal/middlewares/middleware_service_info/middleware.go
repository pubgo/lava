package middleware_service_info

import (
	"context"

	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/grpcutil"
	pbv1 "github.com/pubgo/lava/pkg/proto/lava"
)

func New() lava.Middleware {
	return lava.MiddlewareWrap{
		Name: "service_info",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
				var serverInfo = &pbv1.ServiceInfo{
					Name:     version.Project(),
					Version:  version.Version(),
					Path:     req.Operation(),
					Hostname: running.Hostname,
					Ip:       running.LocalIP,
				}

				if req.Client() {
					req.Header().Set(grpcutil.ClientNameKey, serverInfo.Name)
					req.Header().Set(grpcutil.ClientVersionKey, serverInfo.Version)
					req.Header().Set(grpcutil.ClientPathKey, serverInfo.Path)
					req.Header().Set(grpcutil.ClientHostnameKey, serverInfo.Hostname)
					req.Header().Set(grpcutil.ClientIpKey, serverInfo.Ip)
				} else {
					clientInfo := new(pbv1.ServiceInfo)
					if data := req.Header().Peek(grpcutil.ClientHostnameKey); data != nil {
						clientInfo.Hostname = convert.B2S(data)
					}

					if data := req.Header().Peek(grpcutil.ClientIpKey); data != nil {
						clientInfo.Ip = convert.B2S(data)
					}

					if data := req.Header().Peek(grpcutil.ClientNameKey); data != nil {
						clientInfo.Name = convert.B2S(data)
					}

					if data := req.Header().Peek(grpcutil.ClientVersionKey); data != nil {
						clientInfo.Version = convert.B2S(data)
					}

					if data := req.Header().Peek(grpcutil.ClientPathKey); data != nil {
						clientInfo.Path = convert.B2S(data)
					}

					ctx = lava.CreateCtxWithClientInfo(ctx, clientInfo)
					ctx = lava.CreateCtxWithServerInfo(ctx, serverInfo)
				}

				return next(ctx, req)
			}
		},
	}
}
