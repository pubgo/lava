package jwt

import (
	"context"
	"errors"
	"strings"

	"github.com/pubgo/lava/middleware"
)

const Name = "jwt"

func init() {
	middleware.Register(Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
			v := string(req.Header().Peek("Authorization"))
			if v == "" {
				return ErrNoHeader
			}
			s := strings.Split(v, " ")
			if len(s) != 2 {
				return ErrNoHeader
			}
			to := s[1]
			payload, err := DefaultManager.Verify(to, nil)
			if err != nil {
				return ErrNoHeader
			}
			_ = payload

			return next(ctx, req, resp)
		}
	})
}

//errors
var (
	ErrNoHeader    = errors.New("no authorization in header")
	ErrInvalidAuth = errors.New("invalid authentication")
)
