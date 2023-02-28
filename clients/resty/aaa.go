package resty

import (
	"context"
	"net/url"

	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"
)

const Name = "resty"

// Client http client interface
type Client interface {
	Do(ctx context.Context, req *fasthttp.Request) result.Result[*fasthttp.Response]
	Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Get(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Delete(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Post(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Put(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Patch(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
}
