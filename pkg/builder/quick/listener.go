package quick

import (
	"context"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/pubgo/xerror"
)

var _ net.Listener = (*listener)(nil)

type listener struct {
	conn   *net.UDPConn
	server quic.Listener
}

// Accept waits for and returns the next connection to the listener.
func (s *listener) Accept() (net.Conn, error) {
	session, err := s.server.Accept(context.Background())
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	stream, err := session.AcceptStream(context.Background())
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	conn := &conn{conn: s.conn, session: session, stream: stream}
	return conn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (s *listener) Close() error {
	return s.server.Close()
}

// Addr returns the listener's network address.
func (s *listener) Addr() net.Addr {
	return s.server.Addr()
}
