package signal

import (
	"context"
	"github.com/pubgo/lava/core/runmode"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/logging"
)

const Name = "signal"

func init() {
	flags.Register(&cli.BoolFlag{
		Name:        "block",
		Destination: &runmode.Block,
		Usage:       "Whether block program",
		Value:       runmode.Block,
	})
}

func Block() {
	if !runmode.Block {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	runmode.Signal = <-ch
	logging.S().Infof("signal [%s] trigger", runmode.Signal)
}

func Ctx() context.Context {
	var ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	return ctx
}
