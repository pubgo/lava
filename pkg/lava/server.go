package lava

import (
	"context"
	"net"
)

type Closer interface {
	Close(ctx context.Context) error
}

// Listener provides an interface for starting and stopping the server.
type Listener interface {
	Listen(context.Context, net.Listener) error
}

type Validator interface {
	Validate() error
}
