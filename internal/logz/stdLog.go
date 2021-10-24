package logz

import (
	"io"
	"log"

	"github.com/pubgo/x/byteutil"
	"go.uber.org/zap"
)

// 替换std默认log
func init() {
	On(func(*Log) {
		var stdLog = log.Default()
		*stdLog = *zap.NewStdLog(getName("std"))
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
