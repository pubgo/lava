package logger

import (
	"context"

	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/loggerInter"
)

func GetLog(ctx context.Context) *zap.Logger {
	return loggerInter.GetLog(ctx)
}
