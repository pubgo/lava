package tunnel

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrAuthFailed 表示认证或授权失败。
var ErrAuthFailed = errors.New("tunnel: authentication failed")

// HeaderTunnelToken 是 HTTP 代理可选的 token 头（与 Authorization: Bearer 二选一）。
const HeaderTunnelToken = "X-Tunnel-Token"

// GRPCAuthPrefix 是 gRPC 代理在路由行之后可选的鉴权行前缀。
const GRPCAuthPrefix = "Authorization: Bearer "

// TokenAuthProvider 基于预共享 token 的 AuthProvider 实现。
//
// Agent 注册时在 ServiceInfo.Metadata["auth_token"] 携带 token；
// Gateway 在 Authenticate 阶段校验该 token 是否在允许列表中。
type TokenAuthProvider struct {
	tokens map[string]struct{}
}

// NewTokenAuthProvider 创建 token 鉴权提供者，至少需要一个 token。
func NewTokenAuthProvider(tokens ...string) (*TokenAuthProvider, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("at least one token is required")
	}
	m := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		if t == "" {
			return nil, fmt.Errorf("token must not be empty")
		}
		m[t] = struct{}{}
	}
	return &TokenAuthProvider{tokens: m}, nil
}

func (p *TokenAuthProvider) Authenticate(service *ServiceInfo) error {
	if service == nil {
		return ErrAuthFailed
	}
	token := ""
	if service.Metadata != nil {
		token = service.Metadata["auth_token"]
	}
	if token == "" || !p.validToken(token) {
		return fmt.Errorf("%w: invalid or missing auth_token for service %q", ErrAuthFailed, service.Name)
	}
	return nil
}

func (p *TokenAuthProvider) Authorize(serviceID, clientID string) error {
	if clientID == "" || !p.validToken(clientID) {
		return fmt.Errorf("%w: unauthorized client", ErrAuthFailed)
	}
	return nil
}

func (p *TokenAuthProvider) GenerateToken(service *ServiceInfo) (string, error) {
	if service == nil || service.Metadata == nil {
		return "", ErrAuthFailed
	}
	token := service.Metadata["auth_token"]
	if token == "" || !p.validToken(token) {
		return "", ErrAuthFailed
	}
	return token, nil
}

func (p *TokenAuthProvider) ValidateToken(token string) (*ServiceInfo, error) {
	if !p.validToken(token) {
		return nil, ErrAuthFailed
	}
	return &ServiceInfo{Metadata: map[string]string{"auth_token": token}}, nil
}

func (p *TokenAuthProvider) validToken(token string) bool {
	for allowed := range p.tokens {
		if subtle.ConstantTimeCompare([]byte(token), []byte(allowed)) == 1 {
			return true
		}
	}
	return false
}

// ClientTokenFromRequest 从 HTTP 请求提取客户端 token。
// 优先 X-Tunnel-Token，其次 Authorization: Bearer。
func ClientTokenFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if t := strings.TrimSpace(r.Header.Get(HeaderTunnelToken)); t != "" {
		return t
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

// SetClientTokenHeader 为出站 HTTP 请求设置 token。
func SetClientTokenHeader(r *http.Request, token string) {
	if r == nil || token == "" {
		return
	}
	r.Header.Set("Authorization", "Bearer "+token)
}
