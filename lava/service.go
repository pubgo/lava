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

// Listener provides an interface for starting and stopping the server.
type Listener interface {
	Listen(context.Context, net.Listener) error
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
