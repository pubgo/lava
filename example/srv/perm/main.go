package perm

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/service"

	"github.com/pubgo/lava/example/cmds"
	"github.com/pubgo/lava/example/migrates"
	"github.com/pubgo/lava/example/pkg/proto/permpb"
	"github.com/pubgo/lava/example/srv/perm/grouprpc"
	"github.com/pubgo/lava/example/srv/perm/menurpc"
	"github.com/pubgo/lava/example/srv/perm/orgrpc"
	"github.com/pubgo/lava/example/srv/perm/permrpc"
	"github.com/pubgo/lava/example/srv/perm/rolerpc"
)

var name = "perm"

func NewSrv() service.Service {
	srv := lava.NewSrv(name, "rbac grpc service")
	srv.SubCmd(cmds.Menu())
	srv.Provider(migrates.Migrations)

	permpb.RegisterMenuServiceServer(srv, menurpc.New())
	permpb.RegisterGroupServiceServer(srv, grouprpc.New())
	permpb.RegisterOrgServiceServer(srv, orgrpc.New())
	permpb.RegisterPermServiceServer(srv, permrpc.New())
	permpb.RegisterRoleServiceServer(srv, rolerpc.New())
	return srv
}
