package grpcc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/config"
)

func init() {
	dix.Register(func(c config.Config) map[string]grpc.ClientConnInterface {
		var clients = make(map[string]grpc.ClientConnInterface)
		var cfgMap = make(map[string]*grpcc_config.Cfg)
		xerror.Panic(c.Decode(grpcc_config.Name, cfgMap))
		for name := range cfgMap {
			clients[name] = NewClient(name)
		}
		return clients
	})
}
