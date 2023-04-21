package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
)

const Name = "signal"

func Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	sig := <-ch
	log.Info().Str("signal", sig.String()).Msg("signal trigger")
}

func Ctx() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	return ctx
}
