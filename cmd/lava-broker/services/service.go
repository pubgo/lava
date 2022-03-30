package services

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/core/registry/registry_plugin"
	"github.com/pubgo/lava/debug/debug_plugin"
	"github.com/pubgo/lava/service/gateway"
	"github.com/pubgo/lava/service/service_type"
)

var name = "lava-broker"

func NewService() service_type.Service {
	srv := lava.NewService(name, "lava broker service")

	registry_plugin.Enable(srv)
	debug_plugin.Enable(srv)
	gateway.Enable(srv)

	// rsocket
	//

	return srv
}
