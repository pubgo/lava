package projectinfo

import (
	"context"

	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/httputil"
)

func Middleware() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (lava.Response, error) {
			req.Header().Set(httputil.HeaderXRequestProject, version.Project())
			req.Header().Set(httputil.HeaderXRequestVersion, version.Version())
			return next(ctx, req)
		}
	}
}
