package resty

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/retry"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

// do 创建处理函数
func do(cfg *Config) lava.HandlerFunc {
	client := cfg.Build()
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		r := req.(*requestImpl).req

		defer fasthttp.ReleaseRequest(r.req)

		var err error
		resp := fasthttp.AcquireResponse()

		backoff := retry.NewNoop()
		if r.backoff != nil {
			backoff = r.backoff
		}

		handle := func() error {
			deadline, ok := ctx.Deadline()
			if ok {
				err = client.DoDeadline(r.req, resp, deadline)
			} else {
				err = client.Do(r.req, resp)
			}
			return err
		}
		err = retry.Do(backoff, func(i int) error { return handle() })
		if err != nil {
			return nil, err
		}

		return &responseImpl{resp: resp}, nil
	}
}

// handleHeader 处理请求头
func handleHeader(c *Client, req *Request) {
	header := c.cfg.DefaultHeader
	for k, v := range header {
		req.header.Add(k, v)
	}
}

// handlePath 处理请求路径
func handlePath(c *Client, req *Request) (r result.Result[string]) {
	reqConf := req.cfg

	reqUrl := c.baseUrl.JoinPath(reqConf.Path)
	req.operation = reqUrl.Path

	if v, ok := c.pathTemplates.Load(reqUrl.Path); ok && v != nil {
		return result.Wrap(PathTemplateRun(v.(*fasttemplate.Template), req.params))
	}

	if IsPathTemplate(reqUrl.Path) {
		pathTemplate, err := CreatePathTemplate(reqUrl.Path)
		if err != nil {
			return r.WithErr(err)
		}
		c.pathTemplates.Store(reqUrl.Path, pathTemplate)
	} else {
		return r.WithValue(reqUrl.Path)
	}

	return r
}

// handleContentType 处理内容类型
func handleContentType(c *Client, req *Request) (r result.Result[string]) {
	defaultConf := c.cfg
	reqConf := req.cfg

	defaultContentTypeValue := defaultContentType
	if defaultConf.DefaultContentType != "" {
		defaultContentTypeValue = defaultConf.DefaultContentType
	}

	return result.Wrap(HandleContentType(defaultContentTypeValue, reqConf.ContentType, req.contentType))
}

// doRequest 构建请求
func doRequest(c *Client, req *Request) (rsp result.Result[*fasthttp.Request]) {
	r := fasthttp.AcquireRequest()

	if handleContentType(c, req).
		IfOK(func(val string) {
			r.Header.Set(httputil.HeaderContentType, val)
		}).
		Throw(&rsp) {
		return rsp
	}

	mth := req.cfg.Method
	if mth == "" {
		return rsp.WithErr(fmt.Errorf("http method is empty"))
	}

	r.Header.SetMethod(mth)

	if GetBodyReader(req.body).
		IfOK(func(val []byte) {
			r.SetBodyRaw(val)
		}).
		Throw(&rsp) {
		return rsp
	}

	handleHeader(c, req)

	for k, v := range req.header {
		for i := range v {
			r.Header.Add(k, v[i])
		}
	}

	// enable auth
	if c.cfg.EnableAuth || req.cfg.EnableAuth {
		if c.cfg.BasicToken != "" {
			r.Header.Set(httputil.HeaderAuthorization, "Basic "+c.cfg.BasicToken)
		}

		if c.cfg.JwtToken != "" {
			r.Header.Set(httputil.HeaderAuthorization, "Bearer "+c.cfg.JwtToken)
		}
	}

	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)
	uri.SetScheme(c.baseUrl.Scheme)
	uri.SetHost(c.baseUrl.Host)
	if handlePath(c, req).
		IfOK(func(val string) {
			uri.SetPath(val)
		}).
		Throw(&rsp) {
		return rsp
	}

	if req.query != nil {
		uri.SetQueryString(req.query.Encode())
	}
	r.SetURI(uri)

	if req.backoff == nil {
		if c.backoff != nil {
			req.backoff = c.backoff
		}

		if req.cfg.Backoff != nil {
			req.backoff = req.cfg.Backoff
		}
	}

	return rsp.WithValue(r)
}
