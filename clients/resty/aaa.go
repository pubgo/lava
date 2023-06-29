package resty

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
	maxRedirectsCount  = 16
	DefaultTimeout     = 10 * time.Second
)

const Name = "resty"

type GetContentFunc func() (io.ReadCloser, error)

// Client http client interface
type Client interface {
	Do(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Head(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Get(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Delete(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
	Post(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
	PostForm(ctx context.Context, val url.Values, req *Request) result.Result[*fasthttp.Response]
	Put(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
	Patch(ctx context.Context, data interface{}, req *Request) result.Result[*fasthttp.Response]
}
