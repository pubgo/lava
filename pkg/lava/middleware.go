package lava

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request) (Response, error)

type Middlewares []Middleware

type Middleware interface {
	String() string
	Middleware(next HandlerFunc) HandlerFunc
}

func WithMiddleware(name string, next func(next HandlerFunc) HandlerFunc) MiddlewareWrap {
	return MiddlewareWrap{Name: name, Next: next}
}

type MiddlewareWrap struct {
	Name string
	Next func(next HandlerFunc) HandlerFunc
}

func (m MiddlewareWrap) Middleware(next HandlerFunc) HandlerFunc {
	return m.Next(next)
}

func (m MiddlewareWrap) String() string {
	return m.Name
}

func Chain(middlewares ...Middleware) Middleware {
	return MiddlewareWrap{
		Name: "chain",
		Next: func(next HandlerFunc) HandlerFunc {
			for i := len(middlewares) - 1; i >= 0; i-- {
				if middlewares[i] == nil {
					continue
				}

				next = middlewares[i].Middleware(next)
			}
			return next
		},
	}
}
