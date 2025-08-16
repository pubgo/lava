package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
)

const Name = "signal"

var logger = log.GetLogger(Name)

func getCh() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR1)
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

func WaitRestart(restart func() error) error {
	sigChan := getCh()
	for sig := range sigChan {
		logger.Info().Str("signal", sig.String()).Msg("signal trigger notify")
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			logger.Info().Str("signal", sig.String()).Msg("stop trigger")
			return nil
		case syscall.SIGHUP, syscall.SIGUSR1:
			logger.Info().Str("signal", sig.String()).Msg("restart trigger")
			err := try.Try(restart)
			if err != nil {
				logger.Error().Err(err).Msg("restart error")
				return err
			}
		}
	}
	return nil
}
