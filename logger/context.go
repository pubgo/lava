package logger

import (
	"context"

	"github.com/pubgo/lava/internal/loggerInter"
	"github.com/pubgo/lava/types"
	"go.uber.org/zap"
)

func GetLog(ctx context.Context) Logger {
	return &loggerWrapper{SugaredLogger: loggerInter.GetLog(ctx).Sugar()}
}

var _ Logger = (*loggerWrapper)(nil)

type loggerWrapper struct {
	*zap.SugaredLogger
}

func (t *loggerWrapper) With(args types.M) Logger {
	if args == nil || len(args) == 0 {
		return t
	}

	var fields = make([]interface{}, len(args))
	for k, v := range args {
		fields = append(fields, zap.Any(k, v))
	}
	return &loggerWrapper{SugaredLogger: t.SugaredLogger.With(fields...)}
}
