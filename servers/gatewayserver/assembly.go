package gatewayserver

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/v2/pkg/gateway"
)

// Gateway is the assembled multi-protocol surface backed by one Mux.
//
//	FiberHandler     — HTTP/REST + gRPC-Web (Fiber / fasthttp)
//	WebSocketHandler — WebSocket (net/http only; not Fiber)
//	GRPCServerOptions — native gRPC h2c passthrough
type Gateway struct {
	Mux               *gateway.Mux
	FiberHandler      fiber.Handler
	WebSocketHandler  http.Handler
	GRPCServerOptions []grpc.ServerOption
}

// NewGatewaySurface builds the public entrypoints for an already-registered Mux.
func NewGatewaySurface(mux *gateway.Mux, wsOpts ...gateway.WSOption) *Gateway {
	return &Gateway{
		Mux:               mux,
		FiberHandler:      mux.Handler,
		WebSocketHandler:  mux.WebSocketHandler(wsOpts...),
		GRPCServerOptions: mux.GRPCServerOptions(),
	}
}
