package grpcc

import (
	"google.golang.org/grpc"

	"sync"
)

var clients sync.Map

func NewDirect(service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return GetDefaultCfg().BuildDirect(service, opts...)
}

func GetClient(service string, optFns ...func(service string) []grpc.DialOption) *client {
	var fn = defaultDialOption
	if len(optFns) > 0 {
		fn = optFns[0]
	}

	return &client{
		service: service,
		optFn:   fn,
	}
}
