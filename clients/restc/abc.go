package restc

import (
	"context"
	"net/http"
	"net/url"
)

const Name = "restc"

// Client http clientImpl interface
type Client interface {
	RoundTripper(func(transport http.RoundTripper) http.RoundTripper) error
	Do(ctx context.Context, req *Request) (*Response, error)
	Head(ctx context.Context, url string, opts ...func(req *Request)) (*Response, error)
	Get(ctx context.Context, url string, opts ...func(req *Request)) (*Response, error)
	Delete(ctx context.Context, url string, opts ...func(req *Request)) (*Response, error)
	Post(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error)
	PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *Request)) (*Response, error)
	Put(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error)
	Patch(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error)
}
