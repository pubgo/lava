package middleware_accesslog

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/utils"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/errors/errutil"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/proto/errorpb"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/core/lavacontexts"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/grpcutil"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
)

const Name = "accesslog"

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

		evt := log.NewEvent()
		referer := utils.UnsafeString(req.Header().Referer())
		if referer != "" {
			evt.Str("referer", referer)
		}

		reqId := lavacontexts.GetReqID(ctx)
		evt.Str("request_id", reqId)
		evt.Int64("started_at", now.Unix())
		evt.Str("service", req.Service())
		evt.Str("operation", req.Operation())
		evt.Str("endpoint", req.Endpoint())
		evt.Bool("client", req.Client())
		evt.Str("version", version.Version())
		evt.Str("method", string(req.Header().Method()))
		evt.Str("query", string(req.Header().RequestURI()))

		clientInfo := lavacontexts.GetClientInfo(ctx)
		if clientInfo != nil {
			evt.Str(grpcutil.ClientNameKey, clientInfo.GetName())
			evt.Str(grpcutil.ClientPathKey, clientInfo.GetPath())
		}

		// 错误和panic处理
		defer func() {
			if !generic.IsNil(gErr) {
				logOpts := handleLogOption(req.Header().PeekAll("X-Log-Option"))
				if logOpts["all"] {
					evt.Any("req_body", req.Payload())
					evt.Bytes("req_header", req.Header().Header())
					if rsp != nil {
						evt.Any("rsp_body", rsp.Payload())
						evt.Any("rsp_header", rsp.Header())
					}
				}
			}

			// 持续时间, 毫秒
			latency := time.Since(now)
			evt.Dur("latency", latency)
			evt.Str("user_agent", string(req.Header().UserAgent()))

			if !req.Client() && rsp != nil {
				rsp.Header().Set("Access-Control-Allow-Credentials", "true")
				// rsp.Header().Set("Access-Control-Expose-Headers", "X-Server-Time")
				// rsp.Header().Set("X-Server-Time", fmt.Sprintf("%v", now.Unix()))
				rsp.Header().Set("X-Request-Latency", fmt.Sprintf("%d", latency.Microseconds()))
			}

			// 记录错误日志
			var e *zerolog.Event
			if generic.IsNil(gErr) {
				// Record requests with a timeout of 200 milliseconds
				//if latency > time.Millisecond*200 && !req.Stream() {
				//	e = l.logger.Err(errTimeout).Func(log.WithEvent(evt))
				//} else {
				e = l.logger.Info().Func(log.WithEvent(evt))
				//}
			} else {
				//errors.Debug(gErr)
				e = l.logger.Err(gErr).Func(log.WithEvent(evt))

				pb := errutil.ParseError(gErr)
				{
					if pb.Trace == nil {
						pb.Trace = new(errorpb.ErrTrace)
					}
					pb.Trace.Operation = req.Operation()
					pb.Trace.Service = req.Service()
					pb.Trace.Version = version.Version()
				}

				{
					if pb.Msg != nil {
						pb.Msg = new(errorpb.ErrMsg)
					}
					pb.Msg.Msg = gErr.Error()
					pb.Msg.Detail = fmt.Sprintf("%#v", gErr)
					if pb.Msg.Tags == nil {
						pb.Msg.Tags = make(map[string]string)
					}
				}

				{
					if pb.Code.Message == "" {
						pb.Code.Message = gErr.Error()
					}

					if pb.Code.StatusCode == errorpb.Code_OK {
						log.Warn(ctx).Any("code", pb.Code).Msg("grpc response error with status code is 0")
					}

					if pb.Code.Code == 0 {
						pb.Code.Code = int32(errutil.GrpcCodeToHTTP(codes.Code(pb.Code.StatusCode)))
						pb.Code.StatusCode = errorpb.Code_Internal
					}
				}

				gErr = errutil.ConvertErr2Status(pb).Err()
			}
			e.Msg("record request")
		}()

		// 集成logger到context
		ctx = log.CreateEventCtx(ctx, log.NewEvent().Str("request_id", reqId).Str("operation", req.Operation()))
		return next(ctx, req)
	}
}

func handleLogOption(data [][]byte) (val map[string]bool) {
	if len(data) == 0 {
		val = map[string]bool{}
		return
	}

	val = make(map[string]bool, len(data))
	for i := range data {
		val[convert.B2S(data[i])] = true
	}

	return val
}
