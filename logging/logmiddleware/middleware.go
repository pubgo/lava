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
	errs "github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/core/requestid"
	"github.com/pubgo/lava/service"
)

const Name = "accesslog"

func Middleware(logger log.Logger) service.Middleware {
	logger = logger.WithName(Name)
	return func(next service.HandlerFunc) service.HandlerFunc {
		return func(ctx context.Context, req service.Request, resp service.Response) (gErr error) {
			now := time.Now()

			var evt = log.NewEvent()
			referer := utils.UnsafeString(req.Header().Referer())
			if referer != "" {
				evt.Str("referer", referer)
			}

			var reqId = requestid.GetFromCtx(ctx)
			var tracerID, spanID = tracing.GetFrom(ctx).SpanID()
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
				if c := recover(); c != nil {
					// 获取堆栈信息, 对堆栈信息进行结构化处理
					goroutines, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
					if len(goroutines) != 0 {
						evt.Interface("go_stack", goroutines)
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
				evt.Str("req_body", reqBody)
				evt.Str("rsp_body", rspBody)
				evt.Any("req_header", req.Header())
				evt.Any("rsp_header", resp.Header())

				// 持续时间, 毫秒
				evt.Str("dur", time.Since(now).String())
				evt.Int64("dur_ms", time.Since(now).Milliseconds())

				// 记录错误日志
				if errs.IsNil(gErr) {
					logger.Info().Msg(req.Endpoint())
				} else {
					logger.Err(gErr).Msg(req.Endpoint())
				}
			}()

			if !req.Client() {
				resp.Header().Set("Access-Control-Allow-Credentials", "true")
				resp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
				resp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
			}

			// 集成logger到context
			ctxLog := logger.WithFields(log.Map{"tracerId": tracerID, "spanId": spanID, "requestId": reqId})
			gErr = next(ctxLog.WithCtx(ctx), req, resp)
			var errPb = errors.FromError(gErr)
			errPb.Operation = req.Operation()
			return assert.Must1(status.New(codes.Code(errPb.Code), errPb.ErrMsg).WithDetails(errPb)).Err()
		}
	}
}
