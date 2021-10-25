package basicauth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const HeaderAuth = "Authorization"
const Name = "basic-auth"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
				subject := req.Header().Get(HeaderAuth)
				if subject == "" {
					return ErrNoHeader
				}

				u, p, err := decode(subject)
				if err != nil {
					panic("can not decode base 64:" + err.Error())
					return ErrNoHeader
				}

				err = cfg.Authenticate(u, p)
				if err != nil {
					return ErrNoHeader
				}

				if cfg.Authorize != nil {
					err = cfg.Authorize(u, req)
					if err != nil {
						return ErrNoHeader
					}
				}

				return next(ctx, req, resp)
			}
		},
	})
}

func decode(subject string) (user string, pwd string, err error) {
	parts := strings.Split(subject, " ")
	if len(parts) != 2 {
		return "", "", ErrInvalidAuth

	}
	if parts[0] != "Basic" {
		return "", "", ErrInvalidAuth
	}
	s, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", ErrInvalidBase64
	}

	result := strings.Split(string(s), ":")
	if len(result) != 2 {
		return "", "", ErrInvalidAuth
	}

	return result[0], result[1], nil
}
