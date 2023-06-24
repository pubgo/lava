package middleware_accesslog

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/utils"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/version"
	"github.com/rs/zerolog"

	"github.com/pubgo/lava"
)

const Name = "accesslog"

var errTimeout = errors.New("grpc server response timeout")

func New(logger log.Logger) *LogMiddleware {
	return &LogMiddleware{
		logger: logger.WithName(Name),
	}
}

var _ lava.Middleware = (*LogMiddleware)(nil)

type LogMiddleware struct {
	logger log.Logger
}

func (l LogMiddleware) String() string {
	return Name
}

func (l LogMiddleware) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
		now := time.Now()

		logOpts := handleLogOption(req.Header().PeekAll("X-Log-Option"))

		evt := log.NewEvent()
		referer := utils.UnsafeString(req.Header().Referer())
		if referer != "" {
			evt.Str("referer", referer)
		}

		reqId := lava.GetReqID(ctx)
		evt.Str("request_id", reqId)
		evt.Int64("start_time", now.UnixMicro())
		evt.Str("service", req.Service())
		evt.Str("operation", req.Operation())
		evt.Str("endpoint", req.Endpoint())
		evt.Bool("client", req.Client())
		evt.Str("version", version.Version())

		// 错误和panic处理
		defer func() {
			if !generic.IsNil(gErr) || logOpts["all"] {
				evt.Any("req_body", req.Payload())
				evt.Bytes("req_header", req.Header().Header())
				if rsp != nil {
					evt.Any("rsp_body", rsp.Payload())
					evt.Any("rsp_header", rsp.Header())
				}
			}

			if !req.Client() && rsp != nil {
				rsp.Header().Set("Access-Control-Allow-Credentials", "true")
				rsp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
				rsp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
			}

			// 持续时间, 毫秒
			latency := time.Since(now)
			evt.Dur("latency", latency)
			evt.Str("user_agent", string(req.Header().UserAgent()))

			// 记录错误日志
			var e *zerolog.Event
			if generic.IsNil(gErr) {
				// Record requests with a timeout of 200 milliseconds
				if latency > time.Millisecond*200 {
					e = l.logger.Err(errTimeout).Func(log.WithEvent(evt))
				} else {
					e = l.logger.Info().Func(log.WithEvent(evt))
				}
			} else {
				e = l.logger.Err(gErr).Func(log.WithEvent(evt))
			}
			e.Msg("record request")
		}()

		// 集成logger到context
		ctx = log.CreateEventCtx(ctx, log.NewEvent().Str("request_id", reqId).Str("operation", req.Operation()))
		return next(ctx, req)
	}
}

func handleLogOption(data [][]byte) (val map[string]bool) {
	if data == nil || len(data) == 0 {
		val = map[string]bool{}
		return
	}

	val = make(map[string]bool, len(data))
	for i := range data {
		val[convert.B2S(data[i])] = true
	}

	return val
}
