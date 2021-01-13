package resolver

import (
	"testing"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

func TestBaseResolver(t *testing.T) {
	var br baseResolver
	br.Close()
	br.ResolveNow(resolver.ResolveNowOptions{})
}

type mockedClientConn struct {
	state resolver.State
}

func (m *mockedClientConn) UpdateState(state resolver.State) {
	m.state = state
}

func (m *mockedClientConn) ReportError(err error) {
}

func (m *mockedClientConn) NewAddress(addresses []resolver.Address) {
}

func (m *mockedClientConn) NewServiceConfig(serviceConfig string) {
}

func (m *mockedClientConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return nil
}
