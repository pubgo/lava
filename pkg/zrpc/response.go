package zrpc

import "github.com/pubgo/lava/v2/lava"

var _ lava.Response = (*response)(nil)

type response struct {
	header  lava.ResponseHeader
	payload any
	stream  bool
}

func (r *response) Header() lava.ResponseHeader { return r.header }
func (r *response) Payload() any                 { return r.payload }
func (r *response) Stream() bool                 { return r.stream }
