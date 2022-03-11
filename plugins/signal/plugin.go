package signal

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

const Name = "signal"

var CatchSigpipe = false

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			if CatchSigpipe {
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGPIPE)
				syncx.GoSafe(func() {
					<-sigChan
					logging.L().Warn("Caught SIGPIPE (ignoring all future SIGPIPE)")
					signal.Ignore(syscall.SIGPIPE)
				})
			}
		},
		OnFlags: func() types.Flags {
			return types.Flags{
				&cli.BoolFlag{
					Name:        "catch-sigpipe",
					Destination: &CatchSigpipe,
					Usage:       "catch and ignore SIGPIPE on stdout and stderr if specified",
					Value:       CatchSigpipe,
				},
			}
		},
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
