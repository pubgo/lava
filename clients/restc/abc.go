package restc

import (
	"context"
	"net/url"

	"github.com/valyala/fasthttp"
)

const Name = "restc"

// Client http clientImpl interface
type Client interface {
	Do(ctx context.Context, req *fasthttp.Request) (*fasthttp.Response, error)
	Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	Get(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	Delete(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	Post(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	Put(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
	Patch(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) (*fasthttp.Response, error)
}
