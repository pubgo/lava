package middleware

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, resp Response) error
type Middleware func(next HandlerFunc) HandlerFunc
type Middlewares = []Middleware
