package loggerInter

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

func CtxWithLogger(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, loggerKey{}, logger)
}

func GetLog(ctx context.Context) *zap.Logger {
	var l, ok = ctx.Value(loggerKey{}).(*zap.Logger)
	if ok {
		return l
	}
	return zap.L()
}
