package lava

import (
	"github.com/pubgo/lava/cmd/cmds/running"
	"github.com/urfave/cli/v2"
)

func Run(cmd ...*cli.Command) {
	running.Run(cmd...)
}
