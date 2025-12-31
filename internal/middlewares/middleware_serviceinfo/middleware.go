package middleware_serviceinfo

import (
	"context"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/rs/xid"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/grpcutil"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
)

func New() lava.Middleware {
	return lava.MiddlewareWrap{
		Name: "service_info",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
				reqId := strutil.FirstFnNotEmpty(
					func() string { return lavacontexts.GetReqID(ctx) },
					func() string { return string(req.Header().Peek(httputil.HeaderXRequestID)) },
					func() string { return xid.New().String() },
				)
				ctx = lavacontexts.CreateCtxWithReqID(ctx, reqId)

				defer func() {
					if rsp != nil {
						rsp.Header().Set(httputil.HeaderXRequestID, reqId)
					}
				}()

				serverInfo := &lavapbv1.ServiceInfo{
					Name:     version.Project(),
					Version:  version.Version(),
					Path:     req.Operation(),
					Hostname: running.Hostname,
					Ip:       running.LocalIP,
				}

				clientInfo := new(lavapbv1.ServiceInfo)

				if req.Client() {
					req.Header().Set(grpcutil.ClientNameKey, serverInfo.Name)
					req.Header().Set(grpcutil.ClientVersionKey, serverInfo.Version)
					req.Header().Set(grpcutil.ClientPathKey, serverInfo.Path)
					req.Header().Set(grpcutil.ClientHostnameKey, serverInfo.Hostname)
					req.Header().Set(grpcutil.ClientIpKey, serverInfo.Ip)
				} else {
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
				}

				ctx = lavacontexts.CreateCtxWithClientInfo(ctx, clientInfo)
				ctx = lavacontexts.CreateCtxWithServerInfo(ctx, serverInfo)

				return next(ctx, req)
			}
		},
	}
}
