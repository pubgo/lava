package logRecord

import (
	"context"
	"time"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const name = "logRecord"

func init() {
	plugin.Middleware(name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (err error) {
			var reqId = requestID.GetWith(ctx)
			var log = zap.L().With(zap.String(requestID.Name, reqId))

			now := time.Now()
			var params = make([]zap.Field, 0, 10)
			params = append(params, zap.String("service", req.Service()))
			params = append(params, zap.String("endpoint", req.Endpoint()))

			var respBody interface{}
			err = next(logger.CtxWithLogger(ctx, log), req, func(rsp types.Response) error {
				respBody = rsp.Payload()
				return xerror.Wrap(resp(rsp))
			})

			if err != nil {
				// 根据请求参数决定是否记录请求参数
				params = append(params, zap.Any("req_body", req.Payload()))
				params = append(params, zap.Any("resp_body", respBody))
				params = append(params, zap.Any("header", req.Header()))
				params = append(params, zap.Any("err_msg", err))
			}

			// 微秒
			params = append(params, zap.Int64("duration", time.Since(now).Microseconds()))

			if err != nil {
				log.Error("request record", params...)
				return
			}

			log.Info("request record", params...)

			return
		}
	})
}
