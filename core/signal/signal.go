package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
)

const Name = "signal"

var logger = log.GetLogger(Name)

func getCh() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	return ch
}

func Wait() {
	sig := <-getCh()
	logger.Info().Str("signal", sig.String()).Msg("signal trigger notify")
}

func Context() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := getCh()
	go func() { <-ch; cancel() }()
	return ctx
}
