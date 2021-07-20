package logger

import (
	"context"

	"github.com/pubgo/xlog"
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

type loggerKey struct{}

// WithCtx returns a new context with the provided logger.
func WithCtx(ctx context.Context, log xlog.Xlog) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

// FromCtx retrieves the current logger from the context.
func FromCtx(ctx context.Context) xlog.Xlog {
	logger := ctx.Value(loggerKey{})
	if logger == nil {
		return xlog.GetDefault()
	}

	return logger.(xlog.Xlog)
}
