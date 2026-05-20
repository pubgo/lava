package resty

import (
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

// Response 封装了 fasthttp.Response，提供便捷的方法
type Response struct {
	resp *fasthttp.Response
}

// StatusCode 返回响应状态码
func (r *Response) StatusCode() int {
	return r.resp.StatusCode()
}

// Body 返回响应体
func (r *Response) Body() []byte {
	return r.resp.Body()
}

// String 返回响应体的字符串表示
func (r *Response) String() string {
	return string(r.resp.Body())
}

// JSON 将响应体反序列化为 JSON
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.resp.Body(), v)
}

// Header 返回响应头
func (r *Response) Header() *fasthttp.ResponseHeader {
	return &r.resp.Header
}
