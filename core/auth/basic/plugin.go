package basic

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/pubgo/lava/middleware"
)

const HeaderAuth = "Authorization"
const Name = "basic-auth"

func init() {
	middleware.Register(Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
			subject := string(req.Header().Peek(HeaderAuth))
			if len(subject) == 0 || subject == "" {
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
