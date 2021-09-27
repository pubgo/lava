package quick

import (
	"net"
	"syscall"
	"time"

	"github.com/lucas-clemente/quic-go"
)

var _ net.Conn = (*conn)(nil)

// conn is a generic quic connection implements net.Conn.
type conn struct {
	conn    *net.UDPConn
	session quic.Session
	stream  quic.Stream
}

// Read implements the conn Read method.
func (c *conn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

// Write implements the conn Write method.
func (c *conn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

// LocalAddr returns the local network address.
func (c *conn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

// Close closes the connection.
func (c *conn) Close() error {
	if c.stream == nil {
		return nil
	}

	return c.stream.Close()
}

// SetDeadline sets the deadline associated with the listener. A zero time value disables the deadline.
func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline implements the conn SetReadDeadline method.
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements the conn SetWriteDeadline method.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// SetReadBuffer sets the size of the operating system's receive buffer associated with the connection.
func (c *conn) SetReadBuffer(bytes int) error {
	return c.conn.SetReadBuffer(bytes)
}

// SetWriteBuffer sets the size of the operating system's transmit buffer associated with the connection.
func (c *conn) SetWriteBuffer(bytes int) error {
	return c.conn.SetWriteBuffer(bytes)
}

// SyscallConn returns a raw network connection. This implements the syscall.Conn interface.
func (c *conn) SyscallConn() (syscall.RawConn, error) {
	return c.conn.SyscallConn()
}
