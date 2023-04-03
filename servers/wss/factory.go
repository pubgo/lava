package wss

import "github.com/fasthttp/websocket"

// NewFactory for WSConnection. Saves and passes appropriate option.
func NewFactory(opts ...Option) Factory {
	return &factory{
		savedOptions: opts,
	}
}

type factory struct {
	savedOptions []Option
}

func (f factory) NewConnection(args ConstructorArgs, conn *websocket.Conn, opts ...Option) Connection {
	combinedOptions := append(f.savedOptions, opts...)
	return New(args, conn, combinedOptions...)
}
