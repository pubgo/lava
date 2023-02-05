package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/runmode"
)

const Name = "signal"

func init() {
	defer recovery.Exit()
	flags.Register(&cli.BoolFlag{
		Name:        "block",
		Destination: &runmode.Block,
		Usage:       "Whether block program",
		Value:       runmode.Block,
	})
}

func Wait() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	runmode.Signal = <-ch
	log.Info().Str("signal", runmode.Signal.String()).Msg("signal trigger")
}

func Ctx() context.Context {
	var ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	return ctx
}
