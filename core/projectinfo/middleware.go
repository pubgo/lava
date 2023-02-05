package projectinfo

import (
	"context"

	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/service"
)

func Middleware() service.Middleware {
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, rsp service.Response) (gErr error) {
			req.Header().Set(httpx.HeaderXRequestProject, version.Project())
			req.Header().Set(httpx.HeaderXRequestVersion, version.Version())
			return next(ctx, req, rsp)
		}
	}
}
