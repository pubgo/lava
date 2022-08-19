package restc

import (
	"context"
	"net/url"

	"github.com/pubgo/xerror"
	"github.com/valyala/fasthttp"
)

var defaultClient = xerror.ExitErr(DefaultCfg().Build()).(Client)

func Do(ctx context.Context, req *fasthttp.Request) (*fasthttp.Response, error) {
	return defaultClient.Do(ctx, req)
}

func Get(ctx context.Context, url string, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.Get(ctx, url, requests...)
}

func Delete(ctx context.Context, url string, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.Delete(ctx, url, requests...)
}

func Post(ctx context.Context, url string, data interface{}, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.Post(ctx, url, data, requests...)
}

func PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.PostForm(ctx, url, val, requests...)
}

func Put(ctx context.Context, url string, data interface{}, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.Put(ctx, url, data, requests...)
}

func Patch(ctx context.Context, url string, data interface{}, requests ...func(req *fasthttp.Request)) (*fasthttp.Response, error) {
	return defaultClient.Patch(ctx, url, data, requests...)
}
