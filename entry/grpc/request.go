package grpc

import "github.com/pubgo/lug/encoding"

type rpcRequest struct {
	service     string
	method      string
	contentType string
	cdc         string
	header      map[string]string
	body        []byte
	stream      bool
	payload     interface{}
}

func (r *rpcRequest) ContentType() string {
	return r.contentType
}

func (r *rpcRequest) Service() string {
	return r.service
}

func (r *rpcRequest) Method() string {
	return r.method
}

func (r *rpcRequest) Endpoint() string {
	return r.method
}

func (r *rpcRequest) Codec() string {
	return r.cdc
}

func (r *rpcRequest) Header() map[string]string {
	return r.header
}

func (r *rpcRequest) Read() ([]byte, error) {
	var cdc = encoding.Get(r.cdc)
	if cdc == nil {
		return nil, encoding.ErrNotFound
	}

	return cdc.Marshal(r.payload)
}

func (r *rpcRequest) Stream() bool {
	return r.stream
}

func (r *rpcRequest) Body() interface{} {
	return r.payload
}
