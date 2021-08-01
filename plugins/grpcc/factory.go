package grpcc

import (
	"google.golang.org/grpc"

	"sync"
)

var clients sync.Map

func NewDirect(service string, opts ...func(cfg *Cfg)) (*grpc.ClientConn, error) {
	return GetDefaultCfg(opts...).BuildDirect(service)
}

func GetClient(service string, opts ...func(cfg *Cfg)) *client {
	var fn = func(cfg *Cfg) {}
	if len(opts) > 0 {
		fn = opts[0]
	}
	return &client{service: service, optFn: fn}
}
