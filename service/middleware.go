package service

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, rsp Response) error

type Middleware interface {
	Next(next HandlerFunc) HandlerFunc
}
