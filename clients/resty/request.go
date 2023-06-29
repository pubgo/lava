package resty

import (
	"net/url"
	"regexp"

	"github.com/pubgo/funk/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

var regParam = regexp.MustCompile(`{.+}`)

func NewRequest() *Request {
	return &Request{
		req: fasthttp.AcquireRequest(),
	}
}

type Request struct {
	req          *fasthttp.Request
	pathTemplate *fasttemplate.Template
	err          error
	getBody      GetContentFunc
	operation    string
}

func (req *Request) Err() error {
	return req.err
}

func (req *Request) Copy() *Request {
	var r = fasthttp.AcquireRequest()
	req.req.CopyTo(r)
	return &Request{
		req:          r,
		err:          req.err,
		pathTemplate: req.pathTemplate,
	}
}

// func WithParam(key string, val any) CallOption {
// func WithParams(params map[string]any) CallOption {
// func WithBasicAuth(username, password string) CallOption {
// func WithHeader(key, value string) CallOption {
// func (r *Request) SetURL(url string) *Request {
// func (r *Request) SetFormDataFromValues(data urlpkg.Values) *Request {
// func (r *Request) SetFormData(data map[string]string) *Request {
// func (r *Request) SetFormDataAnyType(data map[string]interface{}) *Request {
// func (r *Request) SetQueryString(query string) *Request {

func (req *Request) WithContentType(contentType string) *Request {

}

func (req *Request) SetQueryValue(params url.Values) *Request {
	req.req.URI().SetQueryString(params.Encode())
	return req
}

func (req *Request) SetPathValue(params map[string]any) *Request {
	if params == nil || len(params) == 0 {
		return req
	}

	if req.pathTemplate == nil {
		return req
	}

	path, err := pathTemplateRun(req.pathTemplate, params)
	if err != nil {
		req.err = err
	} else {
		req.req.URI().SetPath(path)
	}

	return req
}

func (req *Request) WithPath(path string) *Request {
	r := req.Copy()
	if regParam.MatchString(path) {
		pathTemplate, err := fasttemplate.NewTemplate(path, "{", "}")
		assert.Must(err, path)
		r.pathTemplate = pathTemplate
	} else {
		r.req.URI().SetPath(path)
	}
	return r
}
