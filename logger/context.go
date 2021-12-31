package logger

import (
	"context"

	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/loggerInter"
	"github.com/pubgo/lava/types"
)

func GetLog(ctx context.Context) Logger {
	return &loggerWrapper{SugaredLogger: loggerInter.GetLog(ctx).Sugar()}
}

var _ Logger = (*loggerWrapper)(nil)

type loggerWrapper struct {
	*zap.SugaredLogger
}

func (t *loggerWrapper) Depth(depth ...int) Logger {
	if len(depth) == 0 || depth[0] == 0 {
		return t
	}

	return &loggerWrapper{SugaredLogger: t.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(depth[0])).Sugar()}
}

func (t *loggerWrapper) WithErr(err error) Logger {
	if err == nil {
		return t
	}

	return &loggerWrapper{SugaredLogger: t.SugaredLogger.With(zap.String("err", err.Error()), zap.Any("err_stack", err))}
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
