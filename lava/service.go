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
	Close()
}

type Service interface {
	Start()
	Stop()
	Run()
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
