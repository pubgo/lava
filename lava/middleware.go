package lava

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request) (Response, error)
type Middleware func(next HandlerFunc) HandlerFunc

func Chain(m ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			if m[i] == nil {
				continue
			}

			next = m[i](next)
		}
		return next
	}
}
