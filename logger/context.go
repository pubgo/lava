package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

// CreateCtxWith logger wrapper
func CreateCtxWith(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetFrom get log from context
//	从context中获取log, log会带上注入的字段
func GetFrom(ctx context.Context) *zap.Logger {
	// 默认log
	if ctx != nil {
		var log, ok = ctx.Value(loggerKey{}).(*zap.Logger)
		if ok && log != nil {
			return log
		}
	}

	return L()
}
