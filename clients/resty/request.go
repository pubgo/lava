package resty

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/retry"
	"net/http"
	"net/url"
	"regexp"

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

func NewRequest(cfg *RequestConfig) *Request {
	req := &Request{cfg: cfg}

	if regParam.MatchString(cfg.Path) {
		pathTemplate, err := fasttemplate.NewTemplate(cfg.Path, "{", "}")
		assert.Must(err, cfg.Path)
		r.pathTemplate = pathTemplate
	} else {
		r.req.URI().SetPath(path)
	}

	return &Request{
		cfg: cfg,
	}
}

type Request struct {
	cfg          *RequestConfig
	header       http.Header
	cookies      []*http.Cookie
	query        url.Values
	pathTemplate *fasttemplate.Template
	err          error
	body         any
	operation    string
	contentType  string
	retry        retry.Retry
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

func (req *Request) SetBody(body any) *Request {
	req.body = body
	return req
}

func (req *Request) SetQuery(query map[string]string) *Request {
	if query == nil || len(query) == 0 {
		return req
	}
	for k, v := range query {
		req.query.Set(k, v)
	}
	return req
}

func (req *Request) SetQueryString(query string) *Request {
	return req
}

func (req *Request) SetURL(url string) *Request {
	return req
}

func (req *Request) SetBasicAuth(username, password string) *Request {
	return req
}

func (req *Request) SetHeader(key, value string) *Request {
	return req
}

func (req *Request) SetParam(key string, val string) *Request {
	return req
}

func (req *Request) SetParams(params map[string]string) *Request {
	return req
}

func (req *Request) SetContentType(contentType string) *Request {
	req.contentType = contentType
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
