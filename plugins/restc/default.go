package restc

import (
	"context"
	"net/url"

	"github.com/pubgo/xerror"
)

var defaultClient Client

func init() {
	defaultClient = xerror.PanicErr(DefaultCfg().Build()).(Client)
}

func Do(ctx context.Context, req *Request) (*Response, error) { return defaultClient.Do(ctx, req) }

func Get(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Get(ctx, url, requests...)
}

func Delete(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Delete(ctx, url, requests...)
}

func Post(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Post(ctx, url, requests...)
}

func PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.PostForm(ctx, url, val, requests...)
}

func Put(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Put(ctx, url, requests...)
}

func Patch(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Patch(ctx, url, requests...)
}
