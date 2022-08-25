package service

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, rsp Response) error
type Middleware func(next HandlerFunc) HandlerFunc
type Filter func(req Request) bool

type InitMiddleware interface {
	Middlewares() []Middleware
}
