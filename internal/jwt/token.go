package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt"
)

var (
	// ErrMissingHeader means the `Authorization` header was empty.
	ErrMissingHeader = errors.New("the length of the `Authorization` header is zero")
)

// ctx is the context of the JSON web token.
type Context struct {
	UserID   uint64
	Username string
	Uuid     string
	Email    string
	IsAdmin  uint64
}

// secretFunc validates the secret format.
func secretFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Make sure the `alg` is what we except.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(secret), nil
	}
}

// Parse validates the token with the specified secret,
// and returns the context if the token was valid.
func Parse(tokenString string, secret string) (*Context, error) {
	ctx := &Context{}

	// Parse the token.
	token, err := jwt.Parse(tokenString, secretFunc(secret))

	// Parse error.
	if err != nil {
		return ctx, err

		// Read the token if it's valid.
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ctx.UserID = uint64(claims["user_id"].(float64))
		ctx.Username = claims["username"].(string)
		ctx.Uuid = claims["uuid"].(string)
		ctx.Email = claims["email"].(string)
		ctx.IsAdmin = uint64(claims["is_admin"].(float64))
		return ctx, nil

		// Other errors.
	} else {
		return ctx, err
	}
}

// ParseRequest gets the token from the header and
// pass it to the Parse function to parses the token.
func ParseRequest(head http.Header, secret string) (*Context, error) {
	header := head.Get("Authorization")

	if len(header) == 0 {
		return &Context{}, ErrMissingHeader
	}

	var t string
	// Parse the header to get the token part.
	_, err := fmt.Sscanf(header, "Bearer %s", &t)
	if err != nil {
		fmt.Printf("fmt.Sscanf err: %+v", err)
	}
	return Parse(t, secret)
}

// Sign signs the context with the specified secret.
func Sign(ctx context.Context, c Context, secret string) (tokenString string, err error) {
	// The token content.
	// iss: （Issuer）
	// iat: （Issued At）
	// exp: （Expiration Time）
	// aud: （Audience）
	// sub: （Subject）
	// nbf: （Not Before）
	// jti: （JWT ID）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  c.UserID,
		"username": c.Username,
		"uuid":     c.Uuid,
		"email":    c.Email,
		"is_admin": c.IsAdmin,
		"nbf":      time.Now().Unix(),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().AddDate(0, 0, 15).Unix(), //默认 15 天过期
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(secret))

	return
}
