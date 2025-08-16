package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/try"
	"github.com/samber/lo"
)

const Name = "signal"

var logger = log.GetLogger(Name)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR1}

func getCh() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
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

func WaitRestart(restart func() error, close func() error) error {
	sigChan := getCh()
	for sig := range sigChan {
		logger.Info().Str("signal", sig.String()).Msg("signal trigger notify")
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			logger.Info().Str("signal", sig.String()).Msg("stop trigger")
			return try.Try(close)
		case syscall.SIGHUP, syscall.SIGUSR1:
			logger.Info().Str("signal", sig.String()).Msg("restart trigger")
			err := try.Try(restart)
			if err != nil {
				logger.Err(err).Msg("supervisor service restart failed")
			}
			continue
		}
		logger.Error().Msgf("unknown signal: %s, should in signals(%v)", sig, lo.Map(signals, func(s os.Signal, _ int) string { return s.String() }))
	}
	return nil
}
