package logger

import (
	"context"
	"fmt"
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
	"github.com/pubgo/lug/tracing"
	"github.com/pubgo/lug/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: "logger",
		OnInit: func(ent plugin.Entry) {
			var cfg = xlog_config.NewProdConfig()
			if runenv.IsDev() || runenv.IsTest() {
				cfg = xlog_config.NewDevConfig()
				cfg.EncoderConfig.EncodeCaller = consts.Default
			}

			_ = config.Decode("logger", &cfg)
			cfg.Level = runenv.Level
			cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

			// 全局log设置
			var log = xerror.ExitErr(cfg.Build()).(*zap.Logger)
			log = log.Named(runenv.Project)
			zap.ReplaceGlobals(log)
			xerror.Exit(dix.Provider(log))
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
				var reqId = request_id.GetReqID(ctx)
				var log = zap.L().With(
					zap.String(request_id.Name, reqId),
					zap.String("endpoint", req.Endpoint()),
				)

				var span = tracing.FromCtx(ctx)
				xerror.Assert(span == nil, "please init tracing")
				span.SetTag(request_id.Name, reqId)

				now := time.Now()
				var respBody interface{}
				var params []zap.Field
				params = append(params, zap.String("service", req.Service()))
				params = append(params, zap.String("start_time", now.Format(time.RFC3339)))

				defer func() {
					if err1 := recover(); err1 != nil {
						switch err := err1.(type) {
						case error:
							gErr = err
						default:
							gErr = fmt.Errorf("%v", err)
						}
					}

					tracing.SetIfErr(span, gErr)

					if gErr != nil {
						// 根据请求参数决定是否记录请求参数
						params = append(params, zap.Any("req_body", req.Payload()))
						params = append(params, zap.Any("resp_body", respBody))
						params = append(params, zap.Any("header", req.Header()))
						params = append(params, zap.Any("err_msg", gErr))
					}

					// 微秒
					params = append(params, zap.Int64("duration", time.Since(now).Microseconds()))

					defer span.Finish()
					if gErr != nil {
						log.Error("request record", params...)
						return
					}

					log.Info("request record", params...)
				}()

				gErr = next(ctxFromLogger(ctx, log), req, func(rsp types.Response) error {
					respBody = rsp.Payload()
					return xerror.Wrap(resp(rsp))
				})

				return
			}
		},
	})
}
