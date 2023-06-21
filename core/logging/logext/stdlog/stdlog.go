package stdlog

import (
	"io"
	"log"
	"strings"

	logger "github.com/pubgo/funk/log"

	"github.com/pubgo/lava/core/logging"
)

func init() {
	logging.Register("stdLog", SetLogger)
}

// SetLogger 替换std默认log
func SetLogger(logger logger.Logger) {
	stdLog := log.Default()

	// 接管系统默认log
	*stdLog = *log.New(&std{l: logger.WithName("std").WithCallerSkip(3)}, "", 0)
}

var _ io.Writer = (*std)(nil)

type std struct {
	l logger.Logger
}

func (s *std) Write(p []byte) (n int, err error) {
	s.l.Info().Msg(strings.TrimSpace(string(p)))
	return len(p), err
}
