package restc

import (
	"context"
	"net/url"

	"github.com/valyala/fasthttp"
)

func ReleaseResponse(resp *Response) { fasthttp.ReleaseResponse(resp) }

type Request struct {
	*fasthttp.Request
	context.Context
}
type Response = fasthttp.Response
type RequestHeader = fasthttp.RequestHeader
type ResponseHeader = fasthttp.ResponseHeader

// Client http client interface
type Client interface {
	Do(req *Request) (*Response, error)
	Get(url string, requests ...func(req *Request)) (*Response, error)
	Delete(url string, requests ...func(req *Request)) (*Response, error)
	Post(url string, requests ...func(req *Request)) (*Response, error)
	PostForm(url string, val url.Values, requests ...func(req *Request)) (*Response, error)
	Put(url string, requests ...func(req *Request)) (*Response, error)
	Patch(url string, requests ...func(req *Request)) (*Response, error)
}
