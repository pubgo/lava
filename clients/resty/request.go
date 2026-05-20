package resty

import (
	"net/http"
	"net/url"

	"github.com/pubgo/funk/v2/retry"
	"github.com/valyala/fasthttp"
)

// RequestSpec 请求规范结构
type RequestSpec struct {
	Header      map[string]string // 请求头
	Path        string            // 请求路径
	Method      string            // 请求方法
	ContentType string            // 内容类型
	Backoff     retry.Backoff     // 重试策略
	EnableAuth  bool              // 是否启用认证
}

// CreateRequest 创建请求对象
// 返回: 请求对象实例
func (r RequestSpec) CreateRequest() *Request { return NewRequest(&r) }

// NewRequest 创建新的请求对象
// cfg: 请求规范
// 返回: 请求对象实例
func NewRequest(cfg *RequestSpec) *Request {
	r := &Request{
		cfg:    cfg,
		header: make(http.Header),
		query:  make(url.Values),
		params: make(map[string]any),
	}
	return r
}

// Request 请求对象
type Request struct {
	req         *fasthttp.Request // fasthttp 请求对象
	cfg         *RequestSpec      // 请求规范
	header      http.Header       // 请求头
	query       url.Values        // 查询参数
	params      map[string]any    // 路径参数
	operation   string            // 操作名称
	contentType string            // 内容类型
	body        any               // 请求体
	backoff     retry.Backoff     // 重试策略
}

// SetBackoff 设置重试策略
// backoff: 重试策略
// 返回: 请求对象本身（链式调用）
func (req *Request) SetBackoff(backoff retry.Backoff) *Request {
	req.backoff = backoff
	return req
}

// SetBody 设置请求体
// body: 请求体
// 返回: 请求对象本身（链式调用）
func (req *Request) SetBody(body any) *Request {
	req.body = body
	return req
}

// SetQuery 设置查询参数，会覆盖相同键已存在的值。
// query: 查询参数映射。
// 返回: 请求对象本身（链式调用）。
func (req *Request) SetQuery(query map[string]string) *Request {
	if len(query) == 0 {
		return req
	}

	for k, v := range query {
		req.query.Set(k, v)
	}

	return req
}

// AddQuery 添加查询参数，会在相同键下追加值而不是覆盖。
// query: 要追加的查询参数映射。
// 返回: 请求对象本身（链式调用）。
func (req *Request) AddQuery(query map[string]string) *Request {
	if len(query) == 0 {
		return req
	}

	for k, v := range query {
		req.query.Add(k, v)
	}

	return req
}

// AddHeader 添加请求头
// key: 请求头键
// value: 请求头值
// 返回: 请求对象本身（链式调用）
func (req *Request) AddHeader(key, value string) *Request {
	req.header.Add(key, value)
	return req
}

// SetHeader 设置请求头
// key: 请求头键
// value: 请求头值
// 返回: 请求对象本身（链式调用）
func (req *Request) SetHeader(key, value string) *Request {
	req.header.Set(key, value)
	return req
}

// SetParam 设置路径参数
// key: 参数键
// val: 参数值
// 返回: 请求对象本身（链式调用）
func (req *Request) SetParam(key, val string) *Request {
	req.params[key] = val
	return req
}

// SetParams 设置多个路径参数
// params: 路径参数映射
// 返回: 请求对象本身（链式调用）
func (req *Request) SetParams(params map[string]string) *Request {
	for k, v := range params {
		req.params[k] = v
	}
	return req
}

// SetContentType 设置内容类型
// contentType: 内容类型
// 返回: 请求对象本身（链式调用）
func (req *Request) SetContentType(contentType string) *Request {
	req.contentType = FilterFlags(contentType)
	return req
}
