package restc

import (
	"github.com/pubgo/xerror"

	"net/url"
)

var defaultClient Client

func init() {
	defaultClient = xerror.PanicErr(DefaultCfg().Build()).(Client)
}

func Do(req *Request) (*Response, error) { return defaultClient.Do(req) }
func Get(url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Get(url, requests...)
}

func Delete(url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Delete(url, requests...)
}

func Post(url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Post(url, requests...)
}

func PostForm(url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.PostForm(url, val, requests...)
}

func Put(url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Put(url, requests...)
}

func Patch(url string, requests ...func(req *Request)) (*Response, error) {
	return defaultClient.Patch(url, requests...)
}
