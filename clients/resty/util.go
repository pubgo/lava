package resty

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/result"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/httputil"
)

func do(cfg *Config) lava.HandlerFunc {
	var client = cfg.Build()
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var err error
		resp := fasthttp.AcquireResponse()
		deadline, ok := ctx.Deadline()
		if ok {
			err = client.DoDeadline(req.(*requestImpl).req, resp, deadline)
		} else {
			err = client.Do(req.(*requestImpl).req, resp)
		}

		if err != nil {
			return nil, err
		}

		return &responseImpl{resp: resp}, nil
	}
}

func NewRequest() *fasthttp.Request {
	return fasthttp.AcquireRequest()
}

func getBodyReader(rawBody interface{}) ([]byte, error) {
	switch body := rawBody.(type) {
	case nil:
		return nil, nil
	case []byte:
		return body, nil
	case string:
		return convert.StoB(body), nil
	case *bytes.Buffer:
		return body.Bytes(), nil

	// We prioritize *bytes.Reader here because we don't really want to
	// deal with it seeking so want it to match here instead of the
	// io.ReadSeeker case.
	case *bytes.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	// Compat case
	case io.ReadSeeker:
		_, err := body.Seek(0, 0)
		if err != nil {
			return nil, err
		}

		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	// Read all in so we can reset
	case io.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	case url.Values:
		return convert.StoB(body.Encode()), nil

	case json.Marshaler:
		return body.MarshalJSON()

	default:
		bb := bytebufferpool.Get()
		defer bytebufferpool.Put(bb)

		if err := json.NewEncoder(bb).Encode(rawBody); err != nil {
			return nil, err
		}

		return bb.Bytes(), nil
	}
}

// IsRedirect returns true if the status code indicates a redirect.
func IsRedirect(statusCode int) bool {
	return statusCode == http.StatusMovedPermanently ||
		statusCode == http.StatusFound ||
		statusCode == http.StatusSeeOther ||
		statusCode == http.StatusTemporaryRedirect ||
		statusCode == http.StatusPermanentRedirect
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *clientImpl, mth string, url string, data interface{}, opts ...func(req *fasthttp.Request)) (r result.Result[*fasthttp.Response]) {
	body, err := getBodyReader(data)
	if err != nil {
		return r.WithErr(err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	req := fasthttp.AcquireRequest()

	req.Header.Set(httputil.HeaderContentType, defaultContentType)
	req.Header.SetMethod(mth)
	req.Header.SetRequestURI(url)
	req.SetBodyRaw(body)
	if len(opts) > 0 {
		opts[0](req)
	}

	return c.Do(ctx, req)
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
