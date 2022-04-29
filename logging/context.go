package logging

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

var loggerKey = fmt.Sprintf("logging-%s", time.Now())

// CreateCtx create context with logger
func CreateCtx(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLog get log from context
//	从context中获取log, log会带上注入的字段
func GetLog(ctx context.Context) *zap.Logger {
	// 默认log
	if ctx == nil {
		return L()
	}

	if log, ok := ctx.Value(loggerKey).(*zap.Logger); ok && log != nil {
		return log
	}

	return L()
}
