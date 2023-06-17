package lava

import (
	"context"
	"net"
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
