package resty

import (
	"context"
	"net/url"
	"time"

	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"
)

const (
	defaultRetryCount  = 3
	defaultRetryInterval  = 10 * time.Millisecond
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
	maxRedirectsCount  = 16
	DefaultTimeout     = 10 * time.Second
	Name               = "resty"
)

type IClient interface {
	Do(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Head(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Get(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Delete(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Post(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
	PostForm(ctx context.Context, val url.Values, req *Request) result.Result[*fasthttp.Response]
	Put(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
	Patch(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
}
