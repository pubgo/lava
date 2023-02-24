package httpmiddlewares

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/internal/requestid"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/rs/xid"
)

func Recovery() lava.Middleware {
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			rid := strutil.FirstFnNotEmpty(
				func() string { return requestid.GetReqID(ctx) },
				func() string { return string(req.Header().Peek(httputil.HeaderXRequestID)) },
				func() string { return xid.New().String() },
			)
			req.Header().Set(httputil.HeaderXRequestID, rid)

			defer func() {
				if recovery := recover(); recovery != nil {
					debug.PrintStack()
					switch ret := recovery.(type) {
					case error:
						gErr = ret
					case string:
						gErr = errors.New(ret)
					case []byte:
						gErr = errors.New(string(ret))
					default:
						gErr = errors.Parse(ret)
					}

					gErr = errors.WrapStack(gErr)
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
					pb.Id = rid
					pb.Code = uint32(errorpb.Code_Internal)
					pb.Name = "lava.server.panic"
					pb.Reason = gErr.Error()
				}

				gErr = errutil.ConvertErr2Status(pb).Err()
			}()

			ctx = requestid.CreateCtx(ctx, rid)
			rsp, gErr = next(ctx, req)
			return
		}
	}
}
