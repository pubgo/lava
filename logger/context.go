package logger

import (
	"context"

	"github.com/segmentio/ksuid"
)

const (
	xRequestId = "X-Request-Id"
)

func reqID(id string) string {
	if id == "" {
		return ksuid.New().String()
	}

	return id
}

func ReqIDFromCtx(ctx context.Context) string {
	rid, _ := ctx.Value(xRequestId).(string)
	return reqID(rid)
}

func ctxWithReqID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, xRequestId, id)
}
