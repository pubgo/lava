package logging

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

// CreateCtxWith create context with logger
func CreateCtxWith(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// GetLogger get log from context
//	从context中获取log, log会带上注入的字段
func GetLogger(ctx context.Context) *zap.Logger {
	// 默认log
	if ctx == nil {
		return L()
	}

	if log, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok && log != nil {
		return log
	}

	return L()
}
