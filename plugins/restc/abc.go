package restc

import (
	"github.com/valyala/fasthttp"

	"net/url"
)

func ReleaseResponse(resp *Response) { fasthttp.ReleaseResponse(resp) }

type Response = fasthttp.Response
type Request = fasthttp.Request
type RequestHeader = fasthttp.RequestHeader
type ResponseHeader = fasthttp.ResponseHeader

// DoFunc http client do func wrapper
type DoFunc func(req *Request, fn func(resp *Response) error) error

// Middleware http client middleware
type Middleware func(DoFunc) DoFunc

// Client http client interface
type Client interface {
	Do(req *Request) (*Response, error)
	Get(url string, requests ...func(req *Request)) (*Response, error)
	Delete(url string, requests ...func(req *Request)) (*Response, error)
	Post(url string, requests ...func(req *Request)) (*Response, error)
	PostForm(url string, val url.Values, requests ...func(req *Request)) (*Response, error)
	Put(url string, requests ...func(req *Request)) (*Response, error)
	Patch(url string, requests ...func(req *Request)) (*Response, error)
}
