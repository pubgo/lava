package wsconnection

import (
	"context"
	"github.com/gorilla/websocket"
)

// Connection for Websocket Server 2.0
type Connection interface {
	ID() string
	Node() string
	AuthIdentity() string
	Start(ctx context.Context, readMsgHandler func(msgType int, bytes []byte)) error
	Write(msg []byte) error
	Close() error
}

// Factory for Connection.
type Factory interface {
	// NewConnection for WS Server 2.0.
	NewConnection(args ConstructorArgs, conn *websocket.Conn, opts ...Option) Connection
}
