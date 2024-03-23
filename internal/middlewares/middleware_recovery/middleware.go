package middleware_recovery

import (
	"context"
	"fmt"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/httputil"
	"google.golang.org/grpc/codes"
	"runtime/debug"

	"github.com/pubgo/funk/errors"

	"github.com/pubgo/lava/lava"
)

func New() lava.Middleware {
	return lava.MiddlewareWrap{
		Name: "recovery",
		Next: func(next lava.HandlerFunc) lava.HandlerFunc {
			return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
				defer func() {
					if err := errors.Parse(recover()); err != nil {
						debug.PrintStack()
						gErr = errors.WrapStack(err)
					}

					if gErr != nil {
						errors.Debug(gErr)

						defer func() {
							rsp.Header().Set(httputil.HeaderXRequestID, lava.GetReqID(ctx))
							rsp.Header().Set(httputil.HeaderXRequestVersion, version.Version())
							rsp.Header().Set(httputil.HeaderXRequestOperation, req.Operation())
						}()

						pb := errutil.ParseError(gErr)
						if pb.Trace == nil {
							pb.Trace = new(errorpb.ErrTrace)
						}
						pb.Trace.Operation = req.Operation()
						pb.Trace.Service = req.Service()
						pb.Trace.Version = version.Version()

						if pb.Msg != nil {
							pb.Msg = new(errorpb.ErrMsg)
						}
						pb.Msg.Msg = gErr.Error()
						pb.Msg.Detail = fmt.Sprintf("%#v", gErr)
						if pb.Msg.Tags == nil {
							pb.Msg.Tags = make(map[string]string)
						}
						pb.Msg.Tags["reqHeader"] = string(req.Header().Header())

						if pb.Code.Message == "" {
							pb.Code.Message = gErr.Error()
						}

						if pb.Code.Code == 0 {
							pb.Code.StatusCode = errorpb.Code_Internal
							pb.Code.Code = int32(errutil.GrpcCodeToHTTP(codes.Code(uint32(errorpb.Code_Internal))))
						}

						gErr = errutil.ConvertErr2Status(pb).Err()
					}
				}()

				return next(ctx, req)
			}
		},
	}
}
