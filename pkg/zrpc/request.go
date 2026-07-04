package zrpc

import "github.com/pubgo/lava/v2/lava"

var _ lava.Request = (*request)(nil)

type request struct {
	client      bool
	stream      bool
	subject     string
	service     string
	contentType string
	header      lava.RequestHeader
	payload     any
}

func (r *request) Client() bool                { return r.client }
func (r *request) Kind() string                { return lava.RequestKindZrpc }
func (r *request) Stream() bool                { return r.stream }
func (r *request) Service() string             { return r.service }
func (r *request) Operation() string           { return r.subject }
func (r *request) Endpoint() string            { return r.subject }
func (r *request) ContentType() string         { return r.contentType }
func (r *request) Header() lava.RequestHeader { return r.header }
func (r *request) Payload() any                { return r.payload }
