package logmiddleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/tracing"
	"github.com/pubgo/funk/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/service"
)

const Name = "accesslog"

func Middleware(logger log.Logger) service.Middleware {
	logger = logger.WithName(Name)
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request) (rsp service.Response, gErr error) {
			now := time.Now()

			var evt = log.NewEvent()
			referer := utils.UnsafeString(req.Header().Referer())
			if referer != "" {
				evt.Str("referer", referer)
			}

			var reqId = requestid.Ctx(ctx)
			var tracerID, spanID = tracing.Ctx(ctx).SpanID()
			evt.Str("requestId", reqId)
			evt.Str("tracerId", tracerID)
			evt.Str("spanId", spanID)
			evt.Int64("start_time", now.UnixMicro())
			evt.Str("service", req.Service())
			evt.Str("operation", req.Operation())
			evt.Str("endpoint", req.Endpoint())
			evt.Bool("client", req.Client())
			evt.Str("version", version.Version())

			// 错误和panic处理
			defer func() {
				if c := errutil.ParseError(errors.Parse(recover())); c != nil {
					c.Operation = req.Operation()
					c.Name = "lava.middleware.panic"
					c.Code = uint32(errorpb.Code_Internal)
					gErr = errutil.ConvertErr2Status(c).Err()
				}

				// TODO type assert
				reqBody := fmt.Sprintf("%v", req.Payload())
				rspBody := fmt.Sprintf("%v", rsp.Payload())
				evt.Str("req_body", reqBody)
				evt.Str("rsp_body", rspBody)
				evt.Any("req_header", req.Header())
				evt.Any("rsp_header", rsp.Header())

				// 持续时间, 毫秒
				evt.Str("dur", time.Since(now).String())
				evt.Int64("dur_ms", time.Since(now).Milliseconds())

				// 记录错误日志
				if generic.IsNil(gErr) {
					logger.Info().Func(log.WithEvent(evt)).Msg(req.Endpoint())
				} else {
					logger.Err(gErr).Func(log.WithEvent(evt)).Msg(req.Endpoint())
				}
			}()

			if !req.Client() {
				rsp.Header().Set("Access-Control-Allow-Credentials", "true")
				rsp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
				rsp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
			}

			// 集成logger到context
			ctxLog := logger.WithFields(log.Map{"tracerId": tracerID, "spanId": spanID, "requestId": reqId})
			rsp, gErr = next(ctxLog.WithCtx(ctx), req)
			var errPb = errutil.ParseError(gErr)
			errPb.Operation = req.Operation()
			return rsp, assert.Must1(status.New(codes.Code(errPb.Code), errPb.ErrMsg).WithDetails(errPb)).Err()
		}
	}
}
