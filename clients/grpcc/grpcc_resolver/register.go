package grpcc_resolver

import (
	"github.com/pubgo/funk/recovery"
	"google.golang.org/grpc/resolver"
)

func init() {
	defer recovery.Exit()

	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}
