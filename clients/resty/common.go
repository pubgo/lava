package resty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/result"
	"github.com/valyala/fasttemplate"
	"golang.org/x/net/http/httpguts"
)

// IsRedirect 检查状态码是否为重定向
func IsRedirect(statusCode int) bool {
	return statusCode == http.StatusMovedPermanently ||
		statusCode == http.StatusFound ||
		statusCode == http.StatusSeeOther ||
		statusCode == http.StatusTemporaryRedirect ||
		statusCode == http.StatusPermanentRedirect
}

// FilterFlags 过滤内容中的标志
func FilterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}

// ToString 将值转换为字符串
func ToString(v any) string {
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

// PathTemplateRun 运行路径模板
func PathTemplateRun(tpl *fasttemplate.Template, params map[string]any) (string, error) {
	return tpl.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		return w.Write(convert.StoB(ToString(params[tag])))
	})
}

// HeaderGet 获取 HTTP 头
func HeaderGet(h http.Header, key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// HeaderHas 检查 HTTP 头是否存在
func HeaderHas(h http.Header, key string) bool {
	_, ok := h[key]
	return ok
}

// HasPort 检查字符串是否包含端口
func HasPort(s string) bool {
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]")
}

// RemoveEmptyPort 移除空端口
func RemoveEmptyPort(host string) string {
	if HasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

// IsNotToken 检查字符是否不是有效的 HTTP token
func IsNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

// ValidMethod 检查 HTTP 方法是否有效
func ValidMethod(method string) bool {
	return len(method) > 0 && strings.IndexFunc(method, IsNotToken) == -1
}

// ValueOrDefault 返回非空值，否则返回默认值
func ValueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// RequestMethodUsuallyLacksBody 检查 HTTP 方法是否通常不需要请求体
func RequestMethodUsuallyLacksBody(method string) bool {
	switch method {
	case "GET", "HEAD", "DELETE", "OPTIONS", "PROPFIND", "SEARCH":
		return true
	}
	return false
}

// GetBodyReader 获取请求体读取器
func GetBodyReader(rawBody any) (r result.Result[[]byte]) {
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

// CloseBody 关闭请求体
func CloseBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// ErrMissingHost 当请求中没有 Host 或 URL 时返回的错误
var ErrMissingHost = errors.New("http: Request.Write on Request with no Host or URL set")

// ReqWriteExcludeHeader Request.Write 自己处理的头，应该被跳过
var ReqWriteExcludeHeader = map[string]bool{
	"Host":              true, // not in Header map anyway
	"User-Agent":        true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

// HandleContentType 处理内容类型
func HandleContentType(defaultContentType, configContentType, reqContentType string) (string, error) {
	contentType := defaultContentType
	if configContentType != "" {
		contentType = configContentType
	}

	if reqContentType != "" {
		contentType = reqContentType
	}

	if contentType == "" {
		return "", errors.New("content-type header is empty")
	}

	return contentType, nil
}

// CreatePathTemplate 创建路径模板
func CreatePathTemplate(path string) (*fasttemplate.Template, error) {
	return fasttemplate.NewTemplate(path, "{", "}")
}

// IsPathTemplate 检查路径是否包含模板参数
func IsPathTemplate(path string) bool {
	regParam := regexp.MustCompile(`{.+}`)
	return regParam.MatchString(path)
}
