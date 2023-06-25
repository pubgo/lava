package grpcc_resolver

import (
	//"github.com/sercand/kuberesolver/v3"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discoveryBuilder{})
	// resolver.Register(kuberesolver.NewBuilder(nil, "kubernetes"))
}
