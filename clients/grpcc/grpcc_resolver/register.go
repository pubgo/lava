package grpcc_resolver

import (
	"github.com/sercand/kuberesolver/v3"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
	resolver.Register(kuberesolver.NewBuilder(nil, "kubernetes"))
}
