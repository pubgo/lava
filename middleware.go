package lava

import "context"

type HandlerFunc func(ctx context.Context, req Request) (Response, error)

type Middleware interface {
	Middleware(next HandlerFunc) HandlerFunc
}

type MiddlewareWrap func(next HandlerFunc) HandlerFunc

func (m MiddlewareWrap) Middleware(next HandlerFunc) HandlerFunc {
	return m(next)
}

func Chain(m ...Middleware) Middleware {
	return MiddlewareWrap(func(next HandlerFunc) HandlerFunc {
		for i := len(m) - 1; i >= 0; i-- {
			if m[i] == nil {
				continue
			}

			next = m[i].Middleware(next)
		}
		return next
	})
}
