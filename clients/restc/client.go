package restc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	client      *http.Client
	cfg         Cfg
	do          types.MiddleNext
	clientTrace *clientTrace
}

func (c *clientImpl) RoundTripper(f func(transport http.RoundTripper) http.RoundTripper) error {
	if f == nil {
		return nil
	}

	transport := f(c.client.Transport)
	if transport == nil {
		return xerror.New("transport is nil")
	}

	c.client.Transport = f(c.client.Transport)
	return nil
}

func (c *clientImpl) Head(ctx context.Context, url string, opts ...func(req *Request)) (*http.Response, error) {
	return doRequest(ctx, c, http.MethodHead, url, nil, opts...)
}

func (c *clientImpl) Do(ctx context.Context, req *Request) (resp *Response, err error) {
	return resp, c.do(ctx, req, func(res types.Response) error { resp = res.(*response).resp; return nil })
}

func (c *clientImpl) Get(ctx context.Context, url string, opts ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodGet, url, nil, opts...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodDelete, url, data, opts...)
}

func (c *clientImpl) Post(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPost, url, data, opts...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, opts ...func(req *Request)) (*Response, error) {
	var resp, err = doRequest(ctx, c, http.MethodPost, url, nil, func(req *Request) {
		req.req.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(strutil.ToBytes(val.Encode()))), nil
		}

		if len(opts) > 0 {
			opts[0](req)
		}
	})
	return resp, err
}

func (c *clientImpl) Put(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPut, url, data, opts...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, data interface{}, opts ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPatch, url, data, opts...)
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *clientImpl, mth string, url string, data interface{}, opts ...func(req *Request)) (*Response, error) {
	rf, err := getBodyReader(data)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var request = &Request{}

	// Enable trace
	if c.cfg.Trace {
		request.clientTrace = &clientTrace{}
		ctx = request.clientTrace.createContext(ctx)
	}

	req, err := http.NewRequestWithContext(ctx, mth, url, bytes.NewReader(rf))
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}
	req.Header.Set(httpx.HeaderContentType, defaultContentType)

	request.req = req
	if len(opts) > 0 {
		opts[0](request)
	}

	var ct = filterFlags(req.Header.Get(httpx.HeaderContentType))
	var cdc = encoding.GetCdc(ct)
	xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)
	request.ct = ct
	request.cdc = cdc
	request.data = rf

	resp, err := c.Do(ctx, request)
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}

	return resp, nil
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
