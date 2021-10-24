package resolver

import "google.golang.org/grpc/resolver"

func newState(addrs []resolver.Address) resolver.State {
	return resolver.State{Addresses: reshuffle(addrs)}
}
