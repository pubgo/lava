package middleware_recovery

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/version"

	"github.com/pubgo/lava/lava"
)

func New() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			defer func() {
				if err := errors.Parse(recover()); err != nil {
					debug.PrintStack()
					gErr = errors.WrapStack(err)
				}

				if generic.IsNil(gErr) {
					return
				}

				var pb = errutil.ParseError(gErr)
				pb.Operation = req.Operation()
				pb.Service = req.Service()
				pb.Version = version.Version()
				pb.ErrMsg = gErr.Error()
				pb.ErrDetail = []byte(fmt.Sprintf("%#v", gErr))
				if pb.Tags == nil {
					pb.Tags = make(map[string]string)
				}
				pb.Tags["header"] = string(req.Header().Header())

				if pb.Reason == "" {
					pb.Id = lava.GetReqID(ctx)
					pb.Code = errorpb.Code_Internal
					pb.Status = "lava.server.panic"
					pb.Reason = gErr.Error()
				}

				gErr = errutil.ConvertErr2Status(pb).Err()
			}()

			rsp, gErr = next(ctx, req)
			return
		}
	}
}
