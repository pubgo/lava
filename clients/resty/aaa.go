package resty

import (
	"context"
	"net/url"
	"time"

	"github.com/pubgo/funk/result"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava"
)

const Name = "resty"

// Client http client interface
type Client interface {
	Middleware(mm ...lava.Middleware)
	Do(ctx context.Context, req *fasthttp.Request) result.Result[*fasthttp.Response]
	Head(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Get(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Delete(ctx context.Context, url string, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Post(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Put(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
	Patch(ctx context.Context, url string, data interface{}, opts ...func(req *fasthttp.Request)) result.Result[*fasthttp.Response]
}

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
	maxRedirectsCount  = 16
	DefaultTimeout     = 10 * time.Second
)

// e := w.logger.LogEvent().
//		Str("method", methodName).
//		Str("req.ID", reqID.String()).
//		Str("req.timeOut", timeOut.String()).
//		Str("req.service", w.TargetService).
//		Str("req.method", method).
//		Str("req.uri", uri).
//		Str("req.user-agent", w.UserAgent).
//		Str("req.accept", w.Accept).
//		Str("req.content-type", w.ContentType)

// req.SetRequestURI(uri)
//	req.Header.SetContentType(w.ContentType)
//	req.Header.Add("User-Agent", w.UserAgent)
//	req.Header.Add("Accept", w.Accept)
//	req.Header.Add(contracts.ContextKeyRequestID.String(), reqID.String())

// if w.Authentication && len(w.JwtToken) > 0 {
//		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.JwtToken))
//	}
