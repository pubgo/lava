package httputil

import (
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/v2/lava"
)

var (
	_ lava.RequestHeader  = (*FastHTTPRequestHeader)(nil)
	_ lava.ResponseHeader = (*FastHTTPResponseHeader)(nil)
	_ fmt.Stringer        = (*FastHTTPRequestHeader)(nil)
	_ fmt.Stringer        = (*FastHTTPResponseHeader)(nil)
)

// FastHTTPRequestHeader adapts *fasthttp.RequestHeader to lava.RequestHeader.
type FastHTTPRequestHeader struct {
	H *fasthttp.RequestHeader
}

// FastHTTPResponseHeader adapts *fasthttp.ResponseHeader to lava.ResponseHeader.
type FastHTTPResponseHeader struct {
	H *fasthttp.ResponseHeader
}

// NewRequestHeader returns an empty fasthttp-backed request header.
func NewRequestHeader() lava.RequestHeader {
	return &FastHTTPRequestHeader{H: &fasthttp.RequestHeader{}}
}

// NewResponseHeader returns an empty fasthttp-backed response header.
func NewResponseHeader() lava.ResponseHeader {
	return &FastHTTPResponseHeader{H: &fasthttp.ResponseHeader{}}
}

// WrapRequestHeader wraps an existing fasthttp request header.
func WrapRequestHeader(h *fasthttp.RequestHeader) lava.RequestHeader {
	if h == nil {
		return NewRequestHeader()
	}
	return &FastHTTPRequestHeader{H: h}
}

// WrapResponseHeader wraps an existing fasthttp response header.
func WrapResponseHeader(h *fasthttp.ResponseHeader) lava.ResponseHeader {
	if h == nil {
		return NewResponseHeader()
	}
	return &FastHTTPResponseHeader{H: h}
}

// UnwrapRequestHeader returns the underlying fasthttp header when h is a FastHTTPRequestHeader.
func UnwrapRequestHeader(h lava.RequestHeader) *fasthttp.RequestHeader {
	if h == nil {
		return nil
	}
	if wrapped, ok := h.(*FastHTTPRequestHeader); ok {
		return wrapped.H
	}
	return nil
}

// UnwrapResponseHeader returns the underlying fasthttp header when h is a FastHTTPResponseHeader.
func UnwrapResponseHeader(h lava.ResponseHeader) *fasthttp.ResponseHeader {
	if h == nil {
		return nil
	}
	if wrapped, ok := h.(*FastHTTPResponseHeader); ok {
		return wrapped.H
	}
	return nil
}

func (h *FastHTTPRequestHeader) header() *fasthttp.RequestHeader {
	if h == nil || h.H == nil {
		return &fasthttp.RequestHeader{}
	}
	return h.H
}

func (h *FastHTTPResponseHeader) header() *fasthttp.ResponseHeader {
	if h == nil || h.H == nil {
		return &fasthttp.ResponseHeader{}
	}
	return h.H
}

func (h *FastHTTPRequestHeader) Peek(key string) []byte              { return h.header().Peek(key) }
func (h *FastHTTPRequestHeader) PeekAll(key string) [][]byte         { return h.header().PeekAll(key) }
func (h *FastHTTPRequestHeader) Set(key, value string)               { h.header().Set(key, value) }
func (h *FastHTTPRequestHeader) Add(key, value string)               { h.header().Add(key, value) }
func (h *FastHTTPRequestHeader) Method() []byte                      { return h.header().Method() }
func (h *FastHTTPRequestHeader) SetMethod(method string)               { h.header().SetMethod(method) }
func (h *FastHTTPRequestHeader) SetMethodBytes(method []byte)          { h.header().SetMethodBytes(method) }
func (h *FastHTTPRequestHeader) RequestURI() []byte                  { return h.header().RequestURI() }
func (h *FastHTTPRequestHeader) SetRequestURI(uri string)            { h.header().SetRequestURI(uri) }
func (h *FastHTTPRequestHeader) SetRequestURIBytes(uri []byte)       { h.header().SetRequestURIBytes(uri) }
func (h *FastHTTPRequestHeader) ContentType() []byte                 { return h.header().ContentType() }
func (h *FastHTTPRequestHeader) SetContentType(ct string)              { h.header().SetContentType(ct) }
func (h *FastHTTPRequestHeader) SetContentTypeBytes(ct []byte)       { h.header().SetContentTypeBytes(ct) }
func (h *FastHTTPRequestHeader) Referer() []byte                       { return h.header().Referer() }
func (h *FastHTTPRequestHeader) UserAgent() []byte                     { return h.header().UserAgent() }
func (h *FastHTTPRequestHeader) SetCookieBytesKV(key, value []byte)    { h.header().SetCookieBytesKV(key, value) }
func (h *FastHTTPRequestHeader) VisitAll(fn func(key, value []byte)) { h.header().VisitAll(fn) }
func (h *FastHTTPRequestHeader) String() string                      { return h.header().String() }

func (h *FastHTTPResponseHeader) Peek(key string) []byte              { return h.header().Peek(key) }
func (h *FastHTTPResponseHeader) PeekAll(key string) [][]byte         { return h.header().PeekAll(key) }
func (h *FastHTTPResponseHeader) Set(key, value string)               { h.header().Set(key, value) }
func (h *FastHTTPResponseHeader) Add(key, value string)               { h.header().Add(key, value) }
func (h *FastHTTPResponseHeader) VisitAll(fn func(key, value []byte)) { h.header().VisitAll(fn) }
func (h *FastHTTPResponseHeader) String() string                      { return h.header().String() }
