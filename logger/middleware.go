package logger

import (
	"context"
	"time"

	"github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/types"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func Middleware() types.Middleware {
	return func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
			start := time.Now()

			var reqID = ReqIDFromCtx(ctx)
			var respBody interface{}
			ac := make(xlog.M)
			ac["service"] = req.Service()
			ac["method"] = req.Method()
			ac["endpoint"] = req.Endpoint()
			ac["request_id"] = reqID
			ac["start_time"] = start.Format(time.RFC3339)

			defer func() {
				if err := recover(); err != nil || gErr != nil {
					if gErr != nil {
						err = gErr
					}

					// 根据请求参数决定是否记录请求参数
					ac["req_body"] = req.Payload()
					ac["resp_body"] = respBody
					ac["header"] = req.Header()
					ac["error"] = err
				}

				// 微秒
				ac["duration"] = time.Since(start).Microseconds()

				xlog.Info("request log", ac, tracing.TraceIdField(ctx))
			}()

			gErr = next(ctxWithReqID(ctx, reqID), req, func(rsp types.Response) error {
				respBody = rsp.Payload()
				return xerror.Wrap(resp(rsp))
			})

			return gErr
		}
	}
}
