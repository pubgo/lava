package logmiddleware

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/utils"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/result"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/core/tracing"
	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

const Name = "accesslog"

func Middleware(log *logging.Logger) service.Middleware {
	log = log.Named(Name)
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, resp service.Response) (gErr error) {
			now := time.Now()

			// TODO 考虑pool优化
			var params = make([]zap.Field, 0, 20)

			referer := utils.UnsafeString(req.Header().Referer())
			if referer != "" {
				params = append(params, zap.String("referer", referer))
			}

			var reqId = requestid.GetFromCtx(ctx)
			var tracerID, spanID = tracing.GetFrom(ctx).SpanID()

			params = append(params, zap.String("requestId", reqId))
			params = append(params, zap.String("tracerId", tracerID))
			params = append(params, zap.String("spanId", spanID))
			params = append(params, zap.Int64("startTime", now.UnixMicro()))
			params = append(params, zap.String("service", req.Service()))
			params = append(params, zap.String("operation", req.Operation()))
			params = append(params, zap.String("endpoint", req.Endpoint()))
			params = append(params, zap.Bool("client", req.Client()))
			params = append(params, zap.String("version", version.Version()))

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
						gErr = c.(error)
					default:
						gErr = errors.New("lava.middleware.panic").Operation(req.Operation()).Err(fmt.Errorf("%#v", c)).StatusInternal()
					}
				}

				// TODO type assert
				reqBody := fmt.Sprintf("%v", req.Payload())
				rspBody := fmt.Sprintf("%v", resp.Payload())
				params = append(params, zap.String("req_body", reqBody))
				params = append(params, zap.Any("rsp_body", rspBody))

				params = append(params, zap.Any("req_header", req.Header()))
				params = append(params, zap.Any("rsp_header", resp.Header()))

				// 持续时间, 毫秒
				params = append(params, zap.String("duration", time.Since(now).String()))
				params = append(params, zap.Int64("dur_ms", time.Since(now).Milliseconds()))

				// 记录错误日志
				logutil.LogOrErr(log, req.Endpoint(), func() result.Error { return result.WithErr(gErr) }, params...)
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

			gErr = next(ctx, req, resp)
			var errPb = errors.FromError(gErr)
			errPb.Operation = req.Operation()
			return assert.Must1(status.New(codes.Code(errPb.Code), errPb.ErrMsg).WithDetails(errPb)).Err()
		}
	}
}
