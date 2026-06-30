package tunnel

import (
	"crypto/subtle"
	"errors"
	"fmt"
)

// ErrAuthFailed 表示认证或授权失败。
var ErrAuthFailed = errors.New("tunnel: authentication failed")

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
