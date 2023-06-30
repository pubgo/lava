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
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/retry"
	"github.com/valyala/bytebufferpool"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasttemplate"
	"golang.org/x/net/http/httpguts"

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

func handleHeader(c *Client, req *Request) {
	header := c.cfg.DefaultHeader
	if header != nil {
		for k, v := range header {
			req.header.Add(k, v)
		}
	}
}

func handlePath(c *Client, req *Request) (path string, err error) {
	var reqConf = req.cfg

	reqUrl := c.baseUrl.JoinPath(reqConf.Path)
	req.operation = reqUrl.Path
	path = reqUrl.Path

	if v, ok := c.pathTemplates.Load(reqUrl.Path); ok {
		if v != nil {
			path, err = pathTemplateRun(v.(*fasttemplate.Template), req.params)
			if err != nil {
				return
			}
		}
	} else {
		if regParam.MatchString(reqUrl.Path) {
			pathTemplate, err := fasttemplate.NewTemplate(reqUrl.Path, "{", "}")
			if err != nil {
				return "", err
			}
			c.pathTemplates.Store(reqUrl.Path, pathTemplate)
		} else {
			c.pathTemplates.Store(reqUrl.Path, nil)
		}
	}

	return
}

func handleContentType(c *Client, req *Request) (string, error) {
	var defaultConf = c.cfg
	var reqConf = req.cfg

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
		return "", errors.New("context-type header is empty")
	}

	return contentType, nil
}

// doRequest data:[bytes|string|map|struct]
func doRequest(ctx context.Context, c *Client, mth string, body any, req *Request) (rsp result.Result[*fasthttp.Response]) {
	if err := req.Err(); err != nil {
		return rsp.WithErr(err)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	r := fasthttp.AcquireRequest()

	ct, err := handleContentType(c, req)
	if err != nil {
		return rsp.WithErr(err)
	}
	r.Header.Set(httputil.HeaderContentType, ct)

	path, err := handlePath(c, req)
	if err != nil {
		return rsp.WithErr(err)
	}
	r.SetRequestURI(path)

	if mth == "" {
		mth = req.cfg.Method
	}

	r.Header.SetMethod(mth)

	bodyRaw, err := getBodyReader(body)
	if err != nil {
		return rsp.WithErr(err)
	}
	r.SetBodyRaw(bodyRaw)

	handleHeader(c, req)

	for k, v := range req.header {
		for i := range v {
			r.Header.Add(k, v[i])
		}
	}

	// enable auth
	if c.cfg.EnableAuth || req.cfg.EnableAuth {
		if c.cfg.BasicToken != "" {
			r.Header.Set("Authentication", "Basic "+c.cfg.BasicToken)
		}

		if c.cfg.JwtToken != "" {
			r.Header.Set("Authentication", "Bearer "+c.cfg.JwtToken)
		}
	}

	var uri = fasthttp.AcquireURI()
	uri.SetScheme(c.baseUrl.Scheme)
	uri.SetHost(c.baseUrl.Host)
	uri.SetPath(path)
	if req.query != nil {
		uri.SetQueryString(req.query.Encode())
	}
	r.SetURI(uri)

	var retryHandle = c.backoff
	if req.cfg.Retry != nil {
		retryHandle = req.cfg.Retry
	}

	if retryHandle != nil {
		err = retry.New(retryHandle).Do(func(i int) error {
			rsp = c.Do(ctx, req)
			return rsp.Err()
		})
		if err != nil {
			return rsp.WithErr(err)
		}
		return rsp
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
	/*
	     Method         = "OPTIONS"                ; Section 9.2
	                    | "GET"                    ; Section 9.3
	                    | "HEAD"                   ; Section 9.4
	                    | "POST"                   ; Section 9.5
	                    | "PUT"                    ; Section 9.6
	                    | "DELETE"                 ; Section 9.7
	                    | "TRACE"                  ; Section 9.8
	                    | "CONNECT"                ; Section 9.9
	                    | extension-method
	   extension-method = token
	     token          = 1*<any CHAR except CTLs or separators>
	*/
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

func closeRequestBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

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
