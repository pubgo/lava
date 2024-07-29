package grpcc_resolver

import "google.golang.org/grpc/resolver"

func init() {
	resolver.Register(NewDirectBuilder())
	// resolver.Register(&discoveryBuilder{})
	// resolver.Register(kuberesolver.NewBuilder(nil, "kubernetes"))
}
