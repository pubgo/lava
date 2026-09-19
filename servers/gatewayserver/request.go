package gatewayserver

import (
	"github.com/pubgo/lava/v2/pkg/lava"
)

var _ lava.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	service     string
	method      string
	url         string
	contentType string
	header      lava.RequestHeader
	rspHeader   lava.ResponseHeader
	payload     any
	streamed    bool
}

func (r *rpcRequest) Kind() string               { return lava.RequestKindGrpc }
func (r *rpcRequest) Client() bool               { return false }
func (r *rpcRequest) Header() lava.RequestHeader { return r.header }
func (r *rpcRequest) Payload() any               { return r.payload }
func (r *rpcRequest) ContentType() string        { return r.contentType }
func (r *rpcRequest) Service() string            { return r.service }
func (r *rpcRequest) Operation() string          { return r.method }
func (r *rpcRequest) Endpoint() string           { return r.url }
func (r *rpcRequest) Stream() bool               { return r.streamed }
