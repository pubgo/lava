package jwt

import (
	"context"
	"errors"
	"strings"

	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "jwt"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
				v := req.Header().Get("Authorization")
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
		},
	})
}

//errors
var (
	ErrNoHeader    = errors.New("no authorization in header")
	ErrInvalidAuth = errors.New("invalid authentication")
)
