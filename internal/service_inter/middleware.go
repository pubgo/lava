package service_inter

import (
	"context"
)

type HandlerFunc func(ctx context.Context, req Request, resp func(rsp Response) error) error
type Middleware func(next HandlerFunc) HandlerFunc
