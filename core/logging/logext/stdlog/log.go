package stdlog

import (
	"bytes"
	"io"
	"log"

	"github.com/pubgo/funk/convert"
	logger "github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/logging"
)

var evt = logger.NewEvent().Str("ext", "std")

func init() {
	logging.Register("stdLog", SetLogger)
}

// SetLogger 替换std默认log
func SetLogger(logger logger.Logger) {
	stdLog := log.Default()

	// 接管系统默认log
	*stdLog = *log.New(&std{l: logger.WithEvent(evt).WithCallerSkip(3)}, "", 0)
}

var _ io.Writer = (*std)(nil)

type std struct {
	l logger.Logger
}

func (s *std) Write(p []byte) (n int, err error) {
	s.l.Info().Msg(convert.B2S(bytes.TrimSpace(p)))
	return len(p), err
}
