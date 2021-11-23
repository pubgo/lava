package resolver

import "google.golang.org/grpc/resolver"

func newState(addrList []resolver.Address) resolver.State {
	return resolver.State{Addresses: reshuffle(addrList)}
}
