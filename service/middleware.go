package service

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, rsp Response) error
type Middleware func(next HandlerFunc) HandlerFunc
type IMiddleware interface {
	Middlewares() []Middleware
}
