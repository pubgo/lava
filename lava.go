package lava

import (
	"github.com/pubgo/lava/internal/service/grpcs"
	"github.com/pubgo/lava/service"
)

func Run(services ...service.Command) {
	service.Run(services...)
}

func NewSrv(name string, desc ...string) service.Service {
	return grpcs.New(name, desc...)
}
