package lava_net

import (
	"fmt"
	"net"
	"sync"

	"github.com/fatedier/golib/errors"
)

// Custom listener
type CustomListener struct {
	acceptCh chan net.Conn
	closed   bool
	mu       sync.Mutex
}

func NewCustomListener() *CustomListener {
	return &CustomListener{
		acceptCh: make(chan net.Conn, 64),
	}
}

func (l *CustomListener) Accept() (net.Conn, error) {
	conn, ok := <-l.acceptCh
	if !ok {
		return nil, fmt.Errorf("listener closed")
	}
	return conn, nil
}

func (l *CustomListener) PutConn(conn net.Conn) error {
	err := errors.PanicToError(func() {
		select {
		case l.acceptCh <- conn:
		default:
			conn.Close()
		}
	})
	return err
}

func (l *CustomListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.closed {
		close(l.acceptCh)
		l.closed = true
	}
	return nil
}

func (l *CustomListener) Addr() net.Addr {
	return (*net.TCPAddr)(nil)
}
