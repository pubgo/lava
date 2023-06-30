package resty

import (
	"context"
	"time"

	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"
)

const (
	defaultRetryCount    = 3
	defaultRetryInterval = 10 * time.Millisecond
	defaultHTTPTimeout   = 2 * time.Second
	defaultContentType   = "application/json"
	maxRedirectsCount    = 16
	defaultTimeout       = 10 * time.Second
	Name                 = "resty"
)

type IClient interface {
	Do(ctx context.Context, req *Request) result.Result[*fasthttp.Response]
}
