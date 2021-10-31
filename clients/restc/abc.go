package restc

import (
	"context"
	"net/http"
	"net/url"
)

const Name = "restc"

type Request = http.Request
type Response = http.Response

// Client http clientImpl interface
type Client interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	Get(ctx context.Context, url string, requests ...func(req *http.Request)) (*http.Response, error)
	Delete(ctx context.Context, url string, requests ...func(req *http.Request)) (*http.Response, error)
	Post(ctx context.Context, url string, requests ...func(req *http.Request)) (*http.Response, error)
	PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *http.Request)) (*http.Response, error)
	Put(ctx context.Context, url string, requests ...func(req *http.Request)) (*http.Response, error)
	Patch(ctx context.Context, url string, requests ...func(req *http.Request)) (*http.Response, error)
}
