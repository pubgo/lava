package lava

import (
	"github.com/pubgo/lava/cmd/cmds/running"
	"github.com/pubgo/lava/internal/service/grpcs"
	rests "github.com/pubgo/lava/internal/service/web"
	"github.com/pubgo/lava/service"
)

func Run(services ...service.Command) {
	running.Run(services...)
}

func NewSrv(name string, desc ...string) service.Service {
	return grpcs.New(name, desc...)
}

func NewWeb(name string, desc ...string) service.Web {
	return rests.New(name, desc...)
}
