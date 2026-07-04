// Package serverhttp adapts fiber.Ctx to lava.Request and lava.Response
// for HTTP servers (gatewayserver, https).
package serverhttp

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/lava"
)

var (
	_ lava.Request  = (*Request)(nil)
	_ lava.Response = (*Response)(nil)
)

// Request wraps a Fiber context as a lava HTTP request.
type Request struct {
	Ctx fiber.Ctx
}

// NewRequest returns a lava.Request for the given Fiber context.
func NewRequest(ctx fiber.Ctx) lava.Request {
	return &Request{Ctx: ctx}
}

func (r *Request) Kind() string { return lava.RequestKindHttp }

func (r *Request) Operation() string {
	return fmt.Sprintf("%s %s", r.Ctx.Method(), r.Ctx.Route().Path)
}

func (r *Request) Client() bool                { return false }
func (r *Request) Header() *lava.RequestHeader { return &r.Ctx.Request().Header }
func (r *Request) Payload() any                { return r.Ctx.Body() }
func (r *Request) ContentType() string         { return string(r.Ctx.Request().Header.ContentType()) }
func (r *Request) Service() string             { return r.Ctx.Route().Path }
func (r *Request) Endpoint() string            { return string(r.Ctx.Request().RequestURI()) }
func (r *Request) Stream() bool                { return r.Ctx.Request().IsBodyStream() }

// Response wraps a Fiber context as a lava HTTP response.
type Response struct {
	Ctx fiber.Ctx
}

// NewResponse returns a lava.Response for the given Fiber context.
func NewResponse(ctx fiber.Ctx) lava.Response {
	return &Response{Ctx: ctx}
}

func (h *Response) Header() *lava.ResponseHeader { return &h.Ctx.Response().Header }
func (h *Response) Payload() any                 { return h.Ctx.Response().Body() }
func (h *Response) Stream() bool                 { return h.Ctx.Response().IsBodyStream() }
