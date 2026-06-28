package tunnel

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// ServiceURL 构造经 gateway HTTP 代理访问某服务的 URL。
//
// gatewayHTTP 为 gateway HTTP 代理根地址，如 http://gateway:8080；
// serviceName 为已注册服务名；subPath 为服务内路径（可空）。
//
// 示例：ServiceURL("http://gw:8080", "my-api", "/v1/users") → http://gw:8080/my-api/v1/users
func ServiceURL(gatewayHTTP, serviceName, subPath string) string {
	base := strings.TrimRight(gatewayHTTP, "/")
	if subPath == "" {
		subPath = "/"
	}
	if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	p := path.Clean("/" + serviceName + subPath)
	return base + p
}

// ProxyHTTPClient 返回访问 gateway HTTP 代理的 http.Client（标准 URL 路由，无自定义 Transport）。
func ProxyHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{Timeout: timeout}
}

// GetService 经 gateway HTTP 代理 GET 指定服务的子路径。
// token 非空时设置 Authorization: Bearer（gateway 启用 auth 时必填）。
func GetService(client *http.Client, gatewayHTTP, serviceName, subPath, token string) (*http.Response, error) {
	if client == nil {
		client = ProxyHTTPClient(0)
	}
	u, err := url.Parse(ServiceURL(gatewayHTTP, serviceName, subPath))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	SetClientTokenHeader(req, token)
	return client.Do(req)
}
