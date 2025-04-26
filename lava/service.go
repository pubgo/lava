package lava

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Init interface {
	Init()
}

type Close interface {
	Close(ctx context.Context) error
}

type Supervisor interface {
	Serve(ctx context.Context) error
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Server provides an interface for starting and stopping the server.
type Server interface {
	Serve(context.Context, net.Listener) error
}

type Validator interface {
	Validate() error
}

// Initializer ...
type Initializer interface {
	Initialize()
}

type InnerServer struct {
	grpc.ClientConnInterface
}
