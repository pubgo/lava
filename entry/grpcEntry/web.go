package grpcEntry

import (
	"net/http"
	"time"

	"github.com/pubgo/xerror"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

var defaultWebOptions = []grpcweb.Option{
	grpcweb.WithWebsockets(true),
	grpcweb.WithWebsocketPingInterval(time.Second),
	grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
		return true
	}),
	grpcweb.WithCorsForRegisteredEndpointsOnly(false),
	grpcweb.WithOriginFunc(func(origin string) bool {
		return true
	}),
}

// grpcWeb 暂时不上生产
func (g *grpcEntry) grpcWeb(opts ...grpcweb.Option) error {
	if !g.cfg.GrpcWeb {
		return nil
	}

	//var server *grpcweb.WrappedGrpcServer
	//_ = server.IsAcceptableGrpcCorsRequest
	//_ = server.IsGrpcWebRequest

	h := grpcweb.WrapServer(g.srv.Get(), append(defaultWebOptions, opts...)...)
	for _, v := range grpcweb.ListGRPCResources(g.srv.Get()) {
		xerror.Panic(g.gw.Get().HandlePath("POST", v, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			h.ServeHTTP(w, r)
		}))
	}
	return nil
}
