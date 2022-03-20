package jwt

import (
	"context"
	"errors"
	"github.com/pubgo/lava/service/service_type"
	"strings"

	"github.com/pubgo/lava/plugin"
)

const Name = "jwt"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnMiddleware: func(next service_type.MiddleNext) service_type.MiddleNext {
			return func(ctx context.Context, req service_type.Request, resp func(rsp service_type.Response) error) error {
				v := req.Header().Get("Authorization")
				if v[0] == "" {
					return ErrNoHeader
				}
				s := strings.Split(v[0], " ")
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
