package golug_request_id

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-uuid"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

const name = "request_id"
const xRequestId = "X-Request-Id"

func httpRequestId(ctx *fiber.Ctx) error {
	rid := RequestIdFromCtx(ctx.Context())
	if rid != "" {
		return nil
	}

	ctx.Context().SetUserValue(xRequestId, requestId(rid))
	return xerror.Wrap(ctx.Next())
}

func unaryServer(ctx context.Context, info *grpc.UnaryServerInfo) context.Context {
	rid := RequestIdFromCtx(ctx)
	if rid != "" {
		return ctx
	}

	return context.WithValue(ctx, xRequestId, requestId(rid))
}

func streamServer(ss grpc.ServerStream, info *grpc.StreamServerInfo) context.Context {
	rid := RequestIdFromCtx(ss.Context())
	if rid != "" {
		return ss.Context()
	}

	return context.WithValue(ss.Context(), xRequestId, requestId(rid))
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
