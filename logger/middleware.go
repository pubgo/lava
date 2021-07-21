package logger

import (
	"context"
	"github.com/pubgo/lug/types"
	"time"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/tracing"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func Middleware() entry.Middleware {
	return func(next entry.Wrapper) entry.Wrapper {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			var span = tracing.FromCtx(ctx)

			start := time.Now()

			var reqID = ReqIDFromCtx(ctx)
			var respBody interface{}
			ac := make(xlog.M)
			ac["service"] = req.Service()
			ac["method"] = req.Method()
			ac["endpoint"] = req.Endpoint()
			ac["request_id"] = reqID
			ac["trace_id"] = span.GetTraceID()
			ac["receive_time"] = start.Format(time.RFC3339Nano)

			var gErr error
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

				// 毫秒
				ac["cost"] = time.Since(start).Milliseconds()

				xlog.Info("request log", ac)
			}()

			gErr = next(ctxWithReqID(ctx, reqID), req, func(rsp types.Response) error {
				respBody = rsp.Payload()
				return xerror.Wrap(resp(rsp))
			})
			return gErr
		}
	}
}
