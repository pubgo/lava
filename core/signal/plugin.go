package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
)

const Name = "signal"

var logger = log.GetLogger(Name)

func Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	sig := <-ch
	logger.Info().Str("signal", sig.String()).Msg("signal trigger notify")
}
