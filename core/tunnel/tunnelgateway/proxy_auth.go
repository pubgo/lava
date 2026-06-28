package tunnelgateway

import (
	"net/http"

	"github.com/pubgo/lava/v2/core/tunnel"
)

// authorizeClient 校验客户端 token 是否有权访问指定服务。
// authProvider 为 nil 时不校验（开发模式）。
func (g *tunnelGateway) authorizeClient(serviceName, token string) error {
	if g.authProvider == nil {
		return nil
	}
	if token == "" {
		if g.metrics != nil {
			g.metrics.ObserveProxyAuthFailure()
		}
		return tunnel.ErrAuthFailed
	}
	if _, err := g.authProvider.ValidateToken(token); err != nil {
		if g.metrics != nil {
			g.metrics.ObserveProxyAuthFailure()
		}
		return err
	}
	if err := g.authProvider.Authorize(serviceName, token); err != nil {
		if g.metrics != nil {
			g.metrics.ObserveProxyAuthFailure()
		}
		return err
	}
	return nil
}

// writeAuthError 向 HTTP 客户端返回 401。
func writeAuthError(w http.ResponseWriter) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
