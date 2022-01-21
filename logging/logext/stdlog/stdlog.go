package stdlog

import (
	"io"
	"log"

	"github.com/pubgo/x/byteutil"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger"
)

// 替换std默认log
func init() {
	logger.On(func(*logger.Event) {
		var stdLog = log.Default()
		// 接管系统默认log
		*stdLog = *zap.NewStdLog(logger.Component("std").L())
	})
}

var _ io.Writer = (*std)(nil)

type std struct {
	l *zap.Logger
}

func (s std) Write(p []byte) (n int, err error) {
	s.l.Info(byteutil.ToStr(p))
	return len(p), err
}
