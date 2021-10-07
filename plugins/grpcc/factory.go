package grpcc

import (
	"google.golang.org/grpc"

	"github.com/pubgo/lug/consts"
)

func NewDirect(addr string, opts ...func(cfg *Cfg)) (*grpc.ClientConn, error) {
	return getCfg(consts.Default, opts...).BuildDirect(addr)
}

func GetClient(service string, opts ...func(cfg *Cfg)) *Client {
	var fn = func(cfg *Cfg) {}
	if len(opts) > 0 {
		fn = opts[0]
	}
	return &Client{service: service, optFn: fn}
}
