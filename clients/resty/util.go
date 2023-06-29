package resty

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/result"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"

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

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case int:
		return strconv.Itoa(t)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(int64(t), 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(uint64(t), 10)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func pathTemplateRun(tpl *fasttemplate.Template, params map[string]any) (string, error) {
	return tpl.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		return w.Write(convert.StoB(toString(params[tag])))
	})
}

// get is like Get, but key must already be in CanonicalHeaderKey form.
func headerGet(h http.Header, key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// has reports whether h has the provided key defined, even if it's
// set to 0-length slice.
func headerHas(h http.Header, key string) bool {
	_, ok := h[key]
	return ok
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// BasicAuthHeaderValue return the header of basic auth.
func BasicAuthHeaderValue(username, password string) string {
	return "Basic " + basicAuth(username, password)
}
