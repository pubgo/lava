package gid_handler

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/internal/example/grpc/pkg/proto/gidpb"
	"github.com/pubgo/lava/lava"
	"google.golang.org/grpc"
)

var (
	_ lava.GrpcProxy = (*IdProxy)(nil)
)

func NewIdProxy() lava.GrpcProxy {
	return &IdProxy{}
}

type IdProxy struct {
}

func (i IdProxy) Middlewares() []lava.Middleware {
	return lava.Middlewares{
		lava.MiddlewareWrap{
			Next: func(next lava.HandlerFunc) lava.HandlerFunc {
				return func(ctx context.Context, req lava.Request) (lava.Response, error) {
					fmt.Println("proxy-header", req.Header().String())
					return next(ctx, req)
				}
			},
			Name: "proxy",
		},
	}
}

func (i IdProxy) ServiceDesc() *grpc.ServiceDesc { return &gidpb.IdProxy_ServiceDesc }

func (i IdProxy) Proxy() grpc.ClientConnInterface {
	return assert.Must1(grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())))
}
