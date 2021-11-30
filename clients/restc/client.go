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
	client      *http.Client
	cfg         Cfg
	do          types.MiddleNext
	clientTrace *clientTrace
}

func (c *clientImpl) TraceInfo() TraceInfo {
	ct := c.clientTrace

	if ct == nil {
		return TraceInfo{}
	}

	ti := TraceInfo{
		DNSLookup:      ct.dnsDone.Sub(ct.dnsStart),
		TLSHandshake:   ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart),
		ServerTime:     ct.gotFirstResponseByte.Sub(ct.gotConn),
		IsConnReused:   ct.gotConnInfo.Reused,
		IsConnWasIdle:  ct.gotConnInfo.WasIdle,
		ConnIdleTime:   ct.gotConnInfo.IdleTime,
		RequestAttempt: c.cfg.RetryCount,
	}

	// Calculate the total time accordingly,
	// when connection is reused
	if ct.gotConnInfo.Reused {
		ti.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		ti.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		ti.TCPConnTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		ti.ConnTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		ti.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}

	// Capture remote address info when connection is non-nil
	if ct.gotConnInfo.Conn != nil {
		ti.RemoteAddr = ct.gotConnInfo.Conn.RemoteAddr()
	}

	return ti
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
		req.Request.Header.Set(httpx.HeaderContentType, "application/x-www-form-urlencoded")
		req.GetBody = func() (io.ReadCloser, error) {
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
	var body []byte
	switch data.(type) {
	case nil:
		body = nil
	case string:
		body = strutil.ToBytes(data.(string))
	case []byte:
		body = data.([]byte)
	default:
		// TODO 其他类型检测
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
	if len(opts) > 0 {
		opts[0](request)
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
