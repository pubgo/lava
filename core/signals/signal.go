package signals

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/v2/log"
)

const Name = "signals"

var logger = log.GetLogger(Name)

// Inspired by
// https://github.com/kubernetes-sigs/controller-runtime/blob/8499b67e316a03b260c73f92d0380de8cd2e97a1/pkg/manager/signals/signal.go#L25
var onlyOneSignalHandler = make(chan struct{})

// var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGUSR1}
var shutdownSignals = []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill}

func Context() context.Context {
	close(onlyOneSignalHandler)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, shutdownSignals...)
	go func() {
		sig := <-ch
		logger.Info().Msgf("cancelling context, received signal:%s", sig)
		cancel()
		sig = <-ch
		logger.Info().Msgf("os exit, received twice signal:%s", sig)
		os.Exit(1)
	}()

	return ctx
}
