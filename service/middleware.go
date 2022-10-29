package service

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, rsp Response) error
type Middleware func(next HandlerFunc) HandlerFunc

func Chain(m ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}
