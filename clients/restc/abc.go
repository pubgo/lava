package restc

import (
	"context"
	"net/url"

	"github.com/valyala/fasthttp"
)

func ReleaseResponse(resp *Response) { fasthttp.ReleaseResponse(resp) }

type Request = fasthttp.Request
type Response = fasthttp.Response
type RequestHeader = fasthttp.RequestHeader
type ResponseHeader = fasthttp.ResponseHeader

// Client http clientImpl interface
type Client interface {
	Do(ctx context.Context, req *Request) (*Response, error)
	Get(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error)
	Delete(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error)
	Post(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error)
	PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *Request)) (*Response, error)
	Put(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error)
	Patch(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error)
}
