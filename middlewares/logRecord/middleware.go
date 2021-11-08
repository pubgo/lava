package logRecord

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/logger"
	"time"

	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/middlewares/requestID"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/version"
)

const Name = "logRecord"

var logs = logz.New(Name)

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (err error) {
			// TODO 考虑pool优化
			var params = make([]zap.Field, 0, 15)

			referer := types.HeaderGet(req.Header(), httpx.HeaderReferer)
			if referer != "" {
				params = append(params, zap.String("referer", referer))
			}

			origin := types.HeaderGet(req.Header(), httpx.HeaderOrigin)
			if origin != "" {
				params = append(params, zap.String("origin", origin))
			}

			var reqId = requestID.GetWith(ctx)

			now := time.Now()
			params = append(params, zap.String("requestId", reqId))
			params = append(params, zap.Int64("startTime", now.UnixMicro()))
			params = append(params, zap.String("service", req.Service()))
			params = append(params, zap.String("method", req.Method()))
			params = append(params, zap.String("endpoint", req.Endpoint()))
			params = append(params, zap.Bool("client", req.Client()))
			params = append(params, zap.String("version", version.Version))

			var respBody interface{}
			err = next(
				logger.CtxWithLogger(ctx, zap.L().With(zap.String("requestId", reqId))),
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

			if err != nil {
				params = append(params, zap.Any("req_body", req.Payload()))
				params = append(params, zap.Any("resp_body", respBody))
				params = append(params, zap.Any("header", req.Header()))
			}

			// 微秒
			params = append(params, zap.Int64("duration", time.Since(now).Microseconds()))
			logs.Logs(req.Endpoint(), func() error { return err }, params...)
			return
		}
	})
}
