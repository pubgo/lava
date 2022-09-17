package stdlog

import (
	"github.com/pubgo/dix/di"
	"io"
	"log"

	"github.com/pubgo/x/byteutil"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
)

// 替换std默认log
func init() {
	di.Provide(func() logging.ExtLog {
		return func(logger *logging.Logger) {
			var stdLog = log.Default()
			// 接管系统默认log
			*stdLog = *zap.NewStdLog(logging.ModuleLog(logger, "std").L())
		}
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
