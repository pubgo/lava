package lava

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/service/service_type"
)

func Run(desc string, entries ...service_type.Service) {
	service.Run(desc, entries...)
}

func NewService(name string, desc string, plugins ...plugin.Plugin) service_type.Service {
	return service.New(name, desc, plugins...)
}
