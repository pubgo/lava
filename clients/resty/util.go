package resty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/funk/v2/retry"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"golang.org/x/net/http/httpguts"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

func do(cfg *Config) lava.HandlerFunc {
	client := cfg.Build()
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		r := req.(*requestImpl).req

		defer fasthttp.ReleaseRequest(r.req)

		var err error
		resp := fasthttp.AcquireResponse()

		handle := func() error {
			deadline, ok := ctx.Deadline()
			if ok {
				err = client.DoDeadline(r.req, resp, deadline)
			} else {
				err = client.Do(r.req, resp)
			}
			return err
		}

		if r.backoff != nil {
			err = retry.New(r.backoff).Do(func(i int) error { return handle() })
		} else {
			err = handle()
		}

		if err != nil {
			return nil, err
		}

		return &responseImpl{resp: resp}, nil
	}
}

func getBodyReader(rawBody any) (r result.Result[[]byte]) {
	switch body := rawBody.(type) {
	case nil:
		return r
	case *bytes.Buffer:
		return r.WithValue(body.Bytes())
	case []byte:
		return r.WithValue(body)
	case string:
		return r.WithValue(convert.StoB(body))

	// We prioritize *bytes.Reader here because we don't really want to
	// deal with it seeking so want it to match here instead of the
	// io.ReadSeeker case.
	case *bytes.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return r.WithErr(err)
		}
		return r.WithValue(buf)

	// Compat case
	case io.ReadSeeker:
		_, err := body.Seek(0, 0)
		if err != nil {
			return r.WithErr(err)
		}

		buf, err := io.ReadAll(body)
		if err != nil {
			return r.WithErr(err)
		}
		return r.WithValue(buf)

	case url.Values:
		return r.WithValue(convert.StoB(body.Encode()))

	// Read all in so we can reset
	case io.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return r.WithErr(err)
		}
		return r.WithValue(buf)

	case json.Marshaler:
		return result.Wrap(body.MarshalJSON())

	default:
		return result.Wrap(json.Marshal(rawBody))
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

func handleHeader(c *Client, req *Request) {
	header := c.cfg.DefaultHeader
	for k, v := range header {
		req.header.Add(k, v)
	}
}

func handlePath(c *Client, req *Request) (r result.Result[string]) {
	reqConf := req.cfg

	reqUrl := c.baseUrl.JoinPath(reqConf.Path)
	req.operation = reqUrl.Path

	if v, ok := c.pathTemplates.Load(reqUrl.Path); ok && v != nil {
		return result.Wrap(pathTemplateRun(v.(*fasttemplate.Template), req.params))
	} else {
		if regParam.MatchString(reqUrl.Path) {
			pathTemplate, err := fasttemplate.NewTemplate(reqUrl.Path, "{", "}")
			if err != nil {
				return r.WithErr(err)
			}
			c.pathTemplates.Store(reqUrl.Path, pathTemplate)
		} else {
			return r.WithValue(reqUrl.Path)
		}
	}

	return r
}

func handleContentType(c *Client, req *Request) (r result.Result[string]) {
	defaultConf := c.cfg
	reqConf := req.cfg

	contentType := defaultContentType
	if defaultConf.DefaultContentType != "" {
		contentType = defaultConf.DefaultContentType
	}

	if reqConf.ContentType != "" {
		contentType = reqConf.ContentType
	}

	if req.contentType != "" {
		contentType = req.contentType
	}

	if contentType == "" {
		return r.WithErr(errors.New("content-type header is empty"))
	}

	return r.WithValue(contentType)
}

// doRequest data:[bytes|string|map|struct]
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

	if getBodyReader(req.body).
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
		return strconv.FormatInt(t, 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
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

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

func validMethod(method string) bool {
	return len(method) > 0 && strings.IndexFunc(method, isNotToken) == -1
}

func closeBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// requestBodyReadError wraps an error from (*Request).write to indicate
// that the error came from a Read call on the Request.Body.
// This error type should not escape the net/http package to users.
type requestBodyReadError struct{ error }

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// errMissingHost is returned by Write when there is no Host or URL present in
// the Request.
var errMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")

// Headers that Request.Write handles itself and should be skipped.
var reqWriteExcludeHeader = map[string]bool{
	"Host":              true, // not in Header map anyway
	"User-Agent":        true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// requestMethodUsuallyLacksBody reports whether the given request
// method is one that typically does not involve a request body.
// This is used by the Transport (via
// transferWriter.shouldSendChunkedRequestBody) to determine whether
// we try to test-read a byte from a non-nil Request.Body when
// Request.outgoingLength() returns -1. See the comments in
// shouldSendChunkedRequestBody.
func requestMethodUsuallyLacksBody(method string) bool {
	switch method {
	case "GET", "HEAD", "DELETE", "OPTIONS", "PROPFIND", "SEARCH":
		return true
	}
	return false
}
