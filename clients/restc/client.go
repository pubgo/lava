package restc

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	json "github.com/goccy/go-json"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/valyala/bytebufferpool"

	"github.com/pubgo/lava/pkg/encoding"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/types"
)

const (
	defaultRetryCount  = 1
	defaultHTTPTimeout = 2 * time.Second
	defaultContentType = "application/json"
)

var _ Client = (*clientImpl)(nil)

// clientImpl is the Client implementation
type clientImpl struct {
	client *http.Client
	cfg    Cfg
	do     types.MiddleNext
}

func (c *clientImpl) Do(ctx context.Context, req *Request) (resp *Response, err error) {

	return resp, c.do(ctx, req, func(res types.Response) error { resp = res.(*response).resp; return nil })
}

func (c *clientImpl) Get(ctx context.Context, url string, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodGet, url, nil, requests...)
}

func (c *clientImpl) Delete(ctx context.Context, url string, data interface{}, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodDelete, url, data, requests...)
}

func (c *clientImpl) Post(ctx context.Context, url string, data interface{}, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPost, url, data, requests...)
}

func (c *clientImpl) PostForm(ctx context.Context, url string, val url.Values, requests ...func(req *Request)) (*Response, error) {
	var resp, err = doRequest(ctx, c, http.MethodPost, url, nil, func(req *Request) {
		req.Request.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(strutil.ToBytes(val.Encode()))), nil
		}

		if len(requests) > 0 {
			requests[0](req)
		}
	})
	return resp, err
}

func (c *clientImpl) Put(ctx context.Context, url string, data interface{}, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPut, url, data, requests...)
}

func (c *clientImpl) Patch(ctx context.Context, url string, data interface{}, requests ...func(req *Request)) (*Response, error) {
	return doRequest(ctx, c, http.MethodPatch, url, data, requests...)
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *clientImpl, mth string, url string, data interface{}, requests ...func(req *Request)) (*Response, error) {
	var body []byte
	switch data.(type) {
	case nil:
		body = nil
	case string:
		body = strutil.ToBytes(data.(string))
	case []byte:
		body = data.([]byte)
	default:
		bb := bytebufferpool.Get()
		defer bytebufferpool.Put(bb)

		if err := json.NewEncoder(bb).Encode(data); err != nil {
			return nil, err
		}
		body = bb.Bytes()
	}

	if ctx == nil {
		ctx = context.Background()
	}

	var request = &Request{}
	req, err := http.NewRequestWithContext(ctx, mth, url, bytes.NewReader(body))
	if err != nil {
		return nil, xerror.Wrap(err, mth, url)
	}
	req.Header.Set(httpx.HeaderContentType, defaultContentType)

	request.Request = req
	if len(requests) > 0 {
		requests[0](request)
	}

	var ct = filterFlags(req.Header.Get(httpx.HeaderContentType))
	var cdc = encoding.GetCdc(ct)
	xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)
	request.ct = ct
	request.cdc = cdc
	request.data = body

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
