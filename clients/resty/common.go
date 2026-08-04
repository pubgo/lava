package resty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/result"
	"github.com/valyala/fasttemplate"
)

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
