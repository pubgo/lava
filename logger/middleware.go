package logger

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	grpcEntry "github.com/pubgo/lug/entry/grpc"
)

func middleware() func(ctx *fiber.Ctx) error {
	start := time.Now()
	return func(ctx *fiber.Ctx) error {
		var reqID = ReqIDFromCtx(ctx.Context())

		ac := make(xlog.M)
		ac["service"] = ctx.OriginalURL()
		ac["req_id"] = reqID
		ac["receive_time"] = start.Format(time.RFC3339Nano)

		defer func() {
			if err := recover(); err != nil {
				// 根据请求参数决定是否记录请求参数
				ac["body"] = ctx.Request().Body()
				ac["header"] = string(ctx.Request().Header.Header())
				ac["error"] = err
			}

			// 毫秒
			ac["cost"] = time.Since(start).Milliseconds()

			xlog.Info("request log", ac)
		}()

		ctx.Context().SetUserValue(xRequestId, reqID)
		return xerror.Wrap(ctx.Next())
	}
}

func unaryServer() grpc.UnaryServerInterceptor {
	start := time.Now()
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var reqID = ReqIDFromCtx(ctx)

		ac := make(xlog.M)
		ac["service"] = info.FullMethod
		ac["req_id"] = reqID
		ac["receive_time"] = start.Unix()

		defer func() {
			if err := recover(); err != nil {
				// 根据请求参数决定是否记录请求参数
				ac["params"] = req
				var md, _ = metadata.FromIncomingContext(ctx)
				ac["header"] = md
				ac["error"] = err
			}

			// 毫秒
			ac["cost"] = time.Since(start).Milliseconds()

			xlog.Info("record log", ac)
		}()

		ctx = ctxWithReqID(ctx, reqID)
		return handler(ctx, req)
	}
}

func streamServer() grpc.StreamServerInterceptor {
	start := time.Now()
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var ctx = ss.Context()
		var reqID = ReqIDFromCtx(ctx)

		ac := make(xlog.M)
		ac["service"] = info.FullMethod
		ac["req_id"] = ReqIDFromCtx(ctx)
		ac["receive_time"] = start.Format(time.RFC3339Nano)

		defer func() {
			if err := recover(); err != nil {
				// 根据请求参数决定是否记录请求参数
				var md, _ = metadata.FromIncomingContext(ctx)
				ac["header"] = md
				ac["error"] = err
			}

			// 毫秒
			ac["cost"] = time.Since(start).Milliseconds()

			xlog.Info("request log", ac)
		}()

		ctx = ctxWithReqID(ctx, reqID)
		return handler(srv, &grpcEntry.ServerStream{ServerStream: ss, WrappedContext: ctx})
	}
}
