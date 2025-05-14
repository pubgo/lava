package lava

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type Init interface {
	Init()
}

type Close interface {
	Close(ctx context.Context) error
}

type Server interface {
	fmt.Stringer

	// Serve starts the server, no async.
	Serve(ctx context.Context) error
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
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
