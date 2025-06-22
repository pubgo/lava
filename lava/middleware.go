package lava

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

type GrpcGatewayMetadata func(ctx context.Context, req *http.Request, rpcPath string, httpPattern string) metadata.MD

type HandlerFunc func(ctx context.Context, req Request) (Response, error)

type Middlewares []Middleware

type Middleware interface {
	Middleware(next HandlerFunc) HandlerFunc
	String() string
}

type MiddlewareWrap struct {
	Next func(next HandlerFunc) HandlerFunc
	Name string
}

func (m MiddlewareWrap) Middleware(next HandlerFunc) HandlerFunc {
	return m.Next(next)
}

func (m MiddlewareWrap) String() string {
	return m.Name
}

func Chain(m ...Middleware) Middleware {
	return MiddlewareWrap{
		Name: "chain",
		Next: func(next HandlerFunc) HandlerFunc {
			for i := len(m) - 1; i >= 0; i-- {
				if m[i] == nil {
					continue
				}

				next = m[i].Middleware(next)
			}
			return next
		},
	}
}
