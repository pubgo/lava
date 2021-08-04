package grpcc

import (
	"google.golang.org/grpc"
)

func NewDirect(addr string, opts ...func(cfg *Cfg)) (*grpc.ClientConn, error) {
	return GetDefaultCfg(opts...).BuildDirect(addr)
}

func GetClient(service string, opts ...func(cfg *Cfg)) *client {
	var fn = func(cfg *Cfg) {}
	if len(opts) > 0 {
		fn = opts[0]
	}
	return &client{service: service, optFn: fn}
}
