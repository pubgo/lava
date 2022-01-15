package logRecord

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/DataDog/gostackparse"
	"go.uber.org/zap"

	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/logger/logutil"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/requestID"
	"github.com/pubgo/lava/plugins/tracing"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
)

const Name = "logRecord"

var logs = logger.Component(Name)

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (err error) {
			// TODO 考虑pool优化
			var params = make([]zap.Field, 0, 20)

			referer := types.HeaderGet(req.Header(), httpx.HeaderReferer)
			if referer != "" {
				params = append(params, zap.String("referer", referer))
			}

			origin := types.HeaderGet(req.Header(), httpx.HeaderOrigin)
			if origin != "" {
				params = append(params, zap.String("origin", origin))
			}

			var reqId = requestID.GetWith(ctx)
			var tracerID, spanID = tracing.GetFrom(ctx).SpanID()

			now := time.Now()
			params = append(params, zap.String("requestId", reqId))
			params = append(params, zap.String("tracerID", tracerID))
			params = append(params, zap.String("spanID", spanID))
			params = append(params, zap.Int64("startTime", now.UnixMicro()))
			params = append(params, zap.String("service", req.Service()))
			params = append(params, zap.String("operation", req.Operation()))
			params = append(params, zap.String("endpoint", req.Endpoint()))
			params = append(params, zap.Bool("client", req.Client()))
			params = append(params, zap.String("version", version.Version))

			var respBody interface{}

			// 错误和panic处理
			defer func() {
				var stack []byte
				if c := recover(); c != nil {
					// 获取堆栈信息
					stack = debug.Stack()
					// 对堆栈信息进行结构化处理
					goroutines, _ := gostackparse.Parse(bytes.NewReader(stack))
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

				if err != nil {
					params = append(params, zap.Any("req_body", req.Payload()))
					params = append(params, zap.Any("resp_body", respBody))
					params = append(params, zap.Any("header", req.Header()))
				}

				// 持续时间, 微秒
				params = append(params, zap.Int64("duration", time.Since(now).Microseconds()))
				// 记录错误日志
				logutil.LogOrErr(logs.L(), req.Endpoint(), func() error { return err }, params...)
			}()

			err = next(
				// 集成logger到context
				logger.CreateCtxWith(ctx, zap.L().Named(logkey.Service).With(
					zap.String("tracerID", tracerID),
					zap.String("spanID", spanID),
					zap.String("requestId", reqId),
				)),
				req,
				func(rsp types.Response) error {
					respBody = rsp.Payload()
					if !req.Client() {
						rsp.Header().Set("Access-Control-Allow-Origin", origin)
						rsp.Header().Set("Access-Control-Allow-Credentials", "true")
						rsp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
						rsp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
					}
					return resp(rsp)
				})
			return
		}
	})
}
