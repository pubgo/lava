package perm

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/service"

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
	srv.Dix(migrates.Migrations)
	srv.Dix(menurpc.New, grouprpc.New, orgrpc.New, permrpc.New, rolerpc.New)
	srv.Dix(permpb.RegisterMenuServiceServer)
	srv.Dix(permpb.RegisterGroupServiceServer)
	srv.Dix(permpb.RegisterOrgServiceServer)
	srv.Dix(permpb.RegisterPermServiceServer)
	srv.Dix(permpb.RegisterRoleServiceServer)
	return srv
}
