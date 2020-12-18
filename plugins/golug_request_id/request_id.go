package golug_request_id

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-uuid"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

const name = "request_id"
const xRequestId = "X-Request-Id"

func httpRequestId() func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		rid := RequestIdFromCtx(ctx.Context())
		if rid != "" {
			return nil
		}

		ctx.Context().SetUserValue(xRequestId, requestId(rid))
		return xerror.Wrap(ctx.Next())
	}
}

func grpcUnaryServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		rid := RequestIdFromCtx(ctx)
		if rid == "" {
			ctx = context.WithValue(ctx, xRequestId, requestId(rid))
		}

		return handler(ctx, req)
	}
}

func grpcStreamServer() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := golug_grpc.WrapServerStream(ss)

		rid := RequestIdFromCtx(ss.Context())
		if rid == "" {
			wss.WrappedContext = context.WithValue(wss.WrappedContext, xRequestId, requestId(rid))
		}

		return handler(srv, wss)
	}
}

func requestId(rid string) string {
	if rid == "" {
		return xerror.PanicStr(uuid.GenerateUUID())
	}
	return rid
}

func RequestIdFromCtx(ctx context.Context) string {
	rid, ok := ctx.Value(xRequestId).(string)
	if !ok {
		return ""
	}
	return rid
}
