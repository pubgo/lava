package logmiddleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/utils"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/core/middlewares"
	"github.com/pubgo/lava/lava"
)

const Name = "accesslog"

var errTimeout = errors.New("grpc server response timeout")

func Middleware(logger log.Logger) lava.Middleware {
	logger = logger.WithName(Name)
	return func(next lava.HandlerFunc) lava.HandlerFunc {
		return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
			now := time.Now()

			var evt = log.NewEvent()
			referer := utils.UnsafeString(req.Header().Referer())
			if referer != "" {
				evt.Str("referer", referer)
			}

			var reqId = middlewares.RequestID(ctx)
			evt.Str("request_id", reqId)
			evt.Int64("start_time", now.UnixMicro())
			evt.Str("service", req.Service())
			evt.Str("operation", req.Operation())
			evt.Str("endpoint", req.Endpoint())
			evt.Bool("client", req.Client())
			evt.Str("version", version.Version())

			// 错误和panic处理
			defer func() {
				// TODO type assert
				reqBody := fmt.Sprintf("%v", req.Payload())
				rspBody := fmt.Sprintf("%v", rsp.Payload())
				evt.Str("req_body", reqBody)
				evt.Str("rsp_body", rspBody)
				evt.Any("req_header", req.Header())
				evt.Any("rsp_header", rsp.Header())

				// 持续时间, 毫秒
				latency := time.Since(now)
				evt.Dur("latency", latency)
				evt.Str("user_agent", string(req.Header().UserAgent()))

				// 记录错误日志
				if generic.IsNil(gErr) {
					// Record requests with a timeout of 200 milliseconds
					if latency > time.Millisecond*200 {
						logger.Err(errTimeout).Func(log.WithEvent(evt)).Msg(req.Endpoint())
					} else {
						logger.Info().Func(log.WithEvent(evt)).Msg(req.Endpoint())
					}
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
			ctx = logger.WithFields(log.Map{"request_id": reqId}).WithCtx(ctx)
			rsp, gErr = next(ctx, req)
			return
		}
	}
}
