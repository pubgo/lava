package restc

import (
	"context"
	"net/http"
	"net/url"
)

const Name = "restc"

type Response = http.Response

// Client http clientImpl interface
type Client interface {
	RoundTripper(func(transport http.RoundTripper) http.RoundTripper) error
	Do(ctx context.Context, req *Request) (*http.Response, error)
	Head(ctx context.Context, url string, opts ...func(req *Request)) (*http.Response, error)
	Get(ctx context.Context, url string, opts ...func(req *Request)) (*http.Response, error)
	Delete(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*http.Response, error)
	Post(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*http.Response, error)
	PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *Request)) (*http.Response, error)
	Put(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*http.Response, error)
	Patch(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*http.Response, error)
}
