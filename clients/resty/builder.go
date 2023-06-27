package resty

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
)

func NewRequest() *Request {
	return &Request{
		Request: fasthttp.AcquireRequest(),
	}
}

type Request struct {
	*fasthttp.Request
	pathTemplate *fasttemplate.Template
}

func (req *Request) SetPath(path string) *Request {
	req.pathTemplate, err = fasttemplate.NewTemplate(path, "{", "}")
	return req
}
