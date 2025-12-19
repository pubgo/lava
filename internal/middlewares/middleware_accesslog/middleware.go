package middleware_accesslog

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/utils"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/errors/errcode"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/proto/errorpb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/grpcutil"
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

		defer func() {
			if gErr != nil {
				evt.Stringer("req_header", req.Header())
				logOpts := handleLogOption(req.Header())
				if logOpts.EnableAll() {
					evt.Any("req_body", req.Payload())
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
			if gErr == nil {
				// Record requests with a timeout of 200 milliseconds
				//if latency > time.Millisecond*200 && !req.Stream() {
				//	e = l.logger.Err(errTimeout).Func(log.WithEvent(evt))
				//} else {
				e = l.logger.Info().Func(log.WithEvent(evt))
				//}
			} else {
				// errors.Debug(gErr)
				e = l.logger.Err(gErr).Func(log.WithEvent(evt))

				pb := errcode.ParseError(gErr)
				if pb.Message == "" {
					pb.Message = gErr.Error()
				}

				if pb.StatusCode == errorpb.Code_OK {
					log.Warn(ctx).Any("code", pb.Code).Msg("grpc response error with status code is 0")
				}

				if pb.Code == 0 {
					pb.Code = int32(errcode.GrpcCodeToHTTP(codes.Code(pb.StatusCode)))
					pb.StatusCode = errorpb.Code_Internal
				}

				gErr = errcode.ConvertErr2Status(pb).Err()
			}
			e.Msg("record request")
		}()

		// 集成logger到context
		ctx = log.CreateFieldsCtx(ctx, log.Fields{"request_id": reqId, "operation": req.Operation()})
		return next(ctx, req)
	}
}

func handleLogOption(header *lava.RequestHeader) *logOption {
	data := header.PeekAll("X-Log-Option")
	val := make(map[string]bool, len(data))
	for i := range data {
		val[convert.B2S(data[i])] = true
	}

	return &logOption{data: val}
}

type logOption struct {
	data map[string]bool
}

func (opt logOption) EnableAll() bool {
	return opt.data["all"]
}
