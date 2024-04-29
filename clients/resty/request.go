package resty

import (
	"net/http"
	"net/url"
	"regexp"

	"github.com/pubgo/funk/retry"
	"github.com/valyala/fasthttp"
)

var regParam = regexp.MustCompile(`{.+}`)

type RequestConfig struct {
	Header      map[string]string
	Path        string
	Method      string
	ContentType string
	Backoff     retry.Backoff
	EnableAuth  bool
}

func NewRequest(cfg *RequestConfig) *Request {
	r := &Request{cfg: cfg}
	return r
}

type Request struct {
	req         *fasthttp.Request
	cfg         *RequestConfig
	header      http.Header
	query       url.Values
	params      map[string]any
	operation   string
	contentType string
	body        any
	backoff     retry.Backoff
}

func (req *Request) SetBackoff(backoff retry.Backoff) *Request {
	req.backoff = backoff
	return req
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

func (req *Request) AddHeader(key, value string) *Request {
	req.header.Add(key, value)
	return req
}

func (req *Request) SetHeader(key, value string) *Request {
	req.header.Set(key, value)
	return req
}

func (req *Request) SetParam(key, val string) *Request {
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
	req.contentType = filterFlags(contentType)
	return req
}
