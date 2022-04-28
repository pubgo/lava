package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/runtime"
)

const Name = "signal"

func init() {
	flags.Register(&cli.BoolFlag{
		Name:        "block",
		Destination: &runtime.Block,
		Usage:       "Whether block program",
		Value:       runtime.Block,
	})
}

func Block() {
	if !runtime.Block {
		return
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	runtime.Signal = <-ch
	logging.S().Infof("signal [%s] trigger", runtime.Signal)
}

func Ctx() context.Context {
	var ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)
	return ctx
}
