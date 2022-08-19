package lava

import (
	"github.com/pubgo/lava/cmd/cmds/running"
	"github.com/pubgo/lava/internal/service/grpcs"
	"github.com/pubgo/lava/service"
	"github.com/urfave/cli/v2"
)

func Run(srv service.Runtime, cmd ...*cli.Command) {
	running.Run(srv, cmd...)
}

func New() service.Service {
	return grpcs.New()
}
