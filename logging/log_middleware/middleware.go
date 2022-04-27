package log_middleware

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/utils"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"runtime/debug"

	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/plugins/requestID"
	"github.com/pubgo/lava/version"
)

const Name = "logRecord"

func init() {
	module.Register(fx.Invoke(func(log *logging.Logger) {
		log = log.Named(logutil.Names(logkey.Component, Name))
		middleware.Register(Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
			return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
				// TODO 考虑pool优化
				var params = make([]zap.Field, 0, 20)

				referer := utils.UnsafeString(req.Header().Referer())
				if referer != "" {
					params = append(params, zap.String("referer", referer))
				}

				var reqId = requestID.GetWith(ctx)
				var tracerID, spanID = tracing.GetFrom(ctx).SpanID()

				now := time.Now()
				params = append(params, zap.String("requestId", reqId))
				params = append(params, zap.String("tracerId", tracerID))
				params = append(params, zap.String("spanId", spanID))
				params = append(params, zap.Int64("startTime", now.UnixMicro()))
				params = append(params, zap.String("service", req.Service()))
				params = append(params, zap.String("operation", req.Operation()))
				params = append(params, zap.String("endpoint", req.Endpoint()))
				params = append(params, zap.Bool("client", req.Client()))
				params = append(params, zap.String("version", version.Version))

				var respBody interface{}
				var respHeader interface{}
				var err error

				// 错误和panic处理
				defer func() {
					if c := recover(); c != nil {

						// 获取堆栈信息, 对堆栈信息进行结构化处理
						goroutines, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
						if len(goroutines) != 0 {
							params = append(params, zap.Any("stack", goroutines))
						}

						switch c.(type) {
						case error:
							err = c.(error)
						default:
							err = errors.Internal("panic", "service=>%s, endpoint=>%s, msg=>%v", req.Service(), req.Endpoint(), err)
						}
					}

					// TODO type assert
					params = append(params, zap.String("req_body", fmt.Sprintf("%s", req.Payload())))
					params = append(params, zap.Any("resp_body", fmt.Sprintf("%s", respBody)))
					params = append(params, zap.Any("req_header", req.Header()))
					params = append(params, zap.Any("resp_header", respHeader))

					// 持续时间, 微秒
					params = append(params, zap.Int64("duration", time.Since(now).Microseconds()))
					// 记录错误日志
					logutil.LogOrErr(log, req.Endpoint(), func() error { return err }, params...)
				}()

				if !req.Client() {
					resp.Header().Set("Access-Control-Allow-Credentials", "true")
					resp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
					resp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
				}

				// 集成logger到context
				ctx = logging.CreateCtx(ctx, zap.L().Named(logkey.Request).With(
					zap.String("tracerId", tracerID),
					zap.String("spanId", spanID),
					zap.String("requestId", reqId)))

				err = next(ctx, req, resp)
				return err
			}
		})
	}))
}
