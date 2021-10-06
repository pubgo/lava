package logger

import (
	"context"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/plugins/request_id"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/types"
)

const name = "logger"

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent plugin.Entry) {
			var cfg = xlog_config.NewProdConfig()
			if runenv.IsDev() || runenv.IsTest() {
				cfg = xlog_config.NewDevConfig()
				cfg.EncoderConfig.EncodeCaller = consts.Default
			}

			_ = config.Decode(name, &cfg)
			cfg.Level = runenv.Level
			cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

			// 全局log设置
			var log = xerror.PanicErr(cfg.Build()).(*zap.Logger)
			log = log.Named(runenv.Project)
			zap.ReplaceGlobals(log)
			xerror.Exit(dix.Provider(log))
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (err error) {
				var reqId = request_id.GetReqID(ctx)
				var log = zap.L().With(
					zap.String(request_id.Name, reqId),
					//zap.String("endpoint", req.Endpoint()),
				)

				now := time.Now()
				var params = make([]zap.Field, 0, 10)
				params = append(params, zap.String("service", req.Service()))
				params = append(params, zap.String("start_time", now.Format(time.RFC3339)))

				var respBody interface{}
				err = next(ctxWithLogger(ctx, log), req, func(rsp types.Response) error {
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
		},
	})
}
