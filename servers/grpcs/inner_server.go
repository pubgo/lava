package grpcs

import (
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/gateway"
)

// NewInner grpc 服务内部通信
func NewInner(handlers []lava.GrpcRouter, dixMiddlewares []lava.Middleware, metric metrics.Metric, log log.Logger) *lava.InnerServer {
	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	middlewares = append(middlewares, dixMiddlewares...)

	cc := new(inprocgrpc.Channel)
	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)
		cc.RegisterService(h.ServiceDesc(), h)
	}

	cc = cc.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	cc = cc.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))
	return &lava.InnerServer{ClientConnInterface: cc}
}

func NewMux(handlers []lava.GrpcRouter, dixMiddlewares []lava.Middleware, metric metrics.Metric, log log.Logger) *gateway.Mux {
	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	middlewares = append(middlewares, dixMiddlewares...)

	mux := gateway.NewMux()
	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)
		mux.RegisterService(desc, h)
	}

	mux.SetUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	mux.SetStreamInterceptor(handlerStreamMiddle(srvMidMap))
	return mux
}
