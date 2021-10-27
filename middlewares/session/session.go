package session

import "context"

type tokenCtxKey struct{}

// Token represents the session token
type Token struct {
	AuthTime int64                  `json:"auth_time"`
	Issuer   string                 `json:"iss"`
	Audience string                 `json:"aud"`
	Expires  int64                  `json:"exp"`
	IssuedAt int64                  `json:"iat"`
	Subject  string                 `json:"sub,omitempty"`
	UID      string                 `json:"uid,omitempty"`
	Claims   map[string]interface{} `json:"claims"`
}

// NewTokenCtx sets the session token to a given context
func NewTokenCtx(ctx context.Context, token *Token) context.Context {
	ctx = context.WithValue(ctx, tokenCtxKey{}, token)
	return ctx
}

// TokenFromCtx returns a session token
func TokenFromCtx(ctx context.Context) *Token {
	if token, ok := ctx.Value(tokenCtxKey{}).(*Token); ok {
		return token
	}

	return nil
}
