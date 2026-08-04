package tunnelgateway

import (
	"context"
	"net"

	"github.com/pubgo/lava/v2/core/tunnel"
)

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

type fakeSession struct {
	closed bool
	open   tunnel.Stream
	err    error
}

func (s *fakeSession) Close() error { s.closed = true; return nil }
func (s *fakeSession) Open(context.Context) (tunnel.Stream, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.open, nil
}

func (s *fakeSession) OpenWithPriority(ctx context.Context, _ int) (tunnel.Stream, error) {
	return s.Open(ctx)
}
func (s *fakeSession) Accept() (tunnel.Stream, error) { return nil, nil }
func (s *fakeSession) IsClosed() bool                 { return s.closed }
func (s *fakeSession) NumStreams() int                { return 0 }
func (s *fakeSession) LocalAddr() net.Addr            { return fakeAddr("127.0.0.1:1") }
func (s *fakeSession) RemoteAddr() net.Addr           { return fakeAddr("127.0.0.1:2") }

type fakeAuth struct {
	validateErr     error
	authorizeErr    error
	authenticateErr error
}

func (a *fakeAuth) Authenticate(*tunnel.ServiceInfo) error { return a.authenticateErr }
func (a *fakeAuth) Authorize(string, string) error         { return a.authorizeErr }
func (a *fakeAuth) GenerateToken(*tunnel.ServiceInfo) (string, error) {
	return "tok", nil
}

func (a *fakeAuth) ValidateToken(string) (*tunnel.ServiceInfo, error) {
	if a.validateErr != nil {
		return nil, a.validateErr
	}
	return &tunnel.ServiceInfo{Name: "svc"}, nil
}

var (
	_ tunnel.Session      = (*fakeSession)(nil)
	_ tunnel.AuthProvider = (*fakeAuth)(nil)
)
