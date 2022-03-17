package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/x/byteutil"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/types"
)

var _ types.Request = (*rpcRequest)(nil)

type rpcRequest struct {
	handler       grpc.UnaryHandler
	handlerStream grpc.StreamHandler
	stream        grpc.ServerStream
	srv           interface{}
	service       string
	method        string
	url           string
	contentType   string
	header        types.Header
	payload       interface{}
}

func (r *rpcRequest) Kind() string         { return Name }
func (r *rpcRequest) Client() bool         { return false }
func (r *rpcRequest) Header() types.Header { return r.header }
func (r *rpcRequest) Payload() interface{} { return r.payload }
func (r *rpcRequest) ContentType() string  { return r.contentType }
func (r *rpcRequest) Service() string      { return r.service }
func (r *rpcRequest) Operation() string    { return r.method }
func (r *rpcRequest) Endpoint() string     { return r.url }
func (r *rpcRequest) Stream() bool         { return r.stream != nil }

var _ types.Request = (*httpRequest)(nil)

type httpRequest struct {
	ctx    *fiber.Ctx
	header types.Header
}

func (r *httpRequest) Client() bool {
	return false
}

func (r *httpRequest) Header() types.Header {
	return r.header
}

func (r *httpRequest) Payload() interface{} {
	return r.ctx.Body()
}

func (r *httpRequest) Body() ([]byte, error) {
	return r.ctx.Body(), nil
}

func (r *httpRequest) ContentType() string {
	return byteutil.ToStr(r.ctx.Request().Header.ContentType())
}

func (r *httpRequest) Service() string {
	return r.ctx.OriginalURL()
}

func (r *httpRequest) Method() string {
	return r.ctx.Method()
}

func (r *httpRequest) Endpoint() string {
	return r.ctx.OriginalURL()
}

func (r *httpRequest) Codec() string {
	return encoding.cdcMapping[r.ContentType()]
}

func (r *httpRequest) Stream() bool {
	return false
}
