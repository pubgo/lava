package middleware_service

import (
	"context"
	"github.com/pubgo/lava"
	"google.golang.org/grpc/metadata"
)

func New() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			if req.Client() {
				defer func() {
					metadata.AppendToOutgoingContext(ctx, "", "")
				}()
			}
			metadata.FromIncomingContext(ctx)
			return next(ctx, req)
		}
	}
}
