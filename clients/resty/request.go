package resty

import (
	"context"
	"net/http"
	"net/url"
	"regexp"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/retry"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

var regParam = regexp.MustCompile(`{.+}`)

type RequestConfig struct {
	Header      map[string]string
	Cookies     []*http.Cookie
	Operation   string
	Path        string
	Method      string
	ContentType string
	Retry       retry.Retry
}

func NewRequest(cfg RequestConfig) *Request {
	r := &Request{cfg: &cfg}

	if regParam.MatchString(cfg.Path) {
		pathTemplate, err := fasttemplate.NewTemplate(cfg.Path, "{", "}")
		assert.Must(err, cfg.Path)
		r.pathTemplate = pathTemplate
	}

	return r
}

type Request struct {
	cfg          *RequestConfig
	header       http.Header
	cookies      []*http.Cookie
	query        url.Values
	params       map[string]string
	pathTemplate *fasttemplate.Template
	err          error
	body         any
	operation    string
	contentType  string
	retry        retry.Retry
	c            Client
}

func (req *Request) Err() error {
	return req.err
}

func (req *Request) Copy() *Request {
	return &Request{
		cfg:          req.cfg,
		err:          req.err,
		header:       req.header,
		cookies:      req.cookies,
		query:        req.query,
		pathTemplate: req.pathTemplate,
		body:         req.body,
		operation:    req.operation,
		contentType:  req.contentType,
		retry:        req.retry,
	}
}

func (req *Request) SetClient(c Client) *Request {
	req.c = c
	return req
}

func (req *Request) Get(ctx context.Context) result.Result[*fasthttp.Response] {

}

func (req *Request) Delete(ctx context.Context) result.Result[*fasthttp.Response] {

}

func (req *Request) Post(ctx context.Context, body any) result.Result[*fasthttp.Response] {
	return doRequest(ctx, req.c, http.MethodPost, req)
}

func (req *Request) Put(ctx context.Context, body any) result.Result[*fasthttp.Response] {

}

func (req *Request) Patch(ctx context.Context, body any) result.Result[*fasthttp.Response] {

}

func (req *Request) SetBody(body any) *Request {
	req.body = body
	return req
}

func (req *Request) SetQuery(query map[string]string) *Request {
	if query == nil || len(query) == 0 {
		return req
	}

	for k, v := range query {
		req.query.Add(k, v)
	}

	return req
}

func (req *Request) SetQueryString(query string) *Request {
	values, err := url.ParseQuery(query)
	if err != nil {
		req.err = errors.Wrap(err, query)
	} else {
		for k, v := range values {
			for i := range v {
				req.query.Add(k, v[i])
			}
		}
	}

	return req
}

func (req *Request) SetBasicAuth(username, password string) *Request {
	req.header.Set("Authentication", BasicAuthHeaderValue(username, password))
	return req
}

func (req *Request) AddHeader(key, value string) *Request {
	req.header.Add(key, value)
	return req
}

func (req *Request) SetHeader(key, value string) *Request {
	req.header.Set(key, value)
	return req
}

func (req *Request) SetParam(key string, val string) *Request {
	req.params[key] = val
	return req
}

func (req *Request) SetParams(params map[string]string) *Request {
	for k, v := range params {
		req.params[k] = v
	}
	return req
}

func (req *Request) SetContentType(contentType string) *Request {
	req.contentType = contentType
	return req
}
