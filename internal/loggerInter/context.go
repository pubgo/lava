package loggerInter

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

// CtxWithLogger logger wrapper
func CtxWithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLog get log from context
func GetLog(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.L()
	}

	var l, ok = ctx.Value(loggerKey{}).(*zap.Logger)
	if ok {
		return l
	}
	return zap.L()
}
