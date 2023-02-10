package stdlog

import (
	"io"
	"log"

	log2 "github.com/pubgo/funk/log"
	"github.com/pubgo/lava/logging/logext"
)

// New 替换std默认log
func New() logext.ExtLog {
	return func(logger log2.Logger) {
		var stdLog = log.Default()
		// 接管系统默认log
		*stdLog = *log.New(&std{l: logger.WithName("std")}, "", 0)
	}
}

var _ io.Writer = (*std)(nil)

type std struct {
	l log2.Logger
}

func (s *std) Write(p []byte) (n int, err error) {
	s.l.Info().Msg(string(p))
	return len(p), err
}
