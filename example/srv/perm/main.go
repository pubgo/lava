package perm

import (
	"github.com/pubgo/lava"
	"github.com/pubgo/lava/service"

	"github.com/pubgo/lava/example/pkg/proto/permpb"
	"github.com/pubgo/lava/example/srv/perm/group"
	"github.com/pubgo/lava/example/srv/perm/menu"
	"github.com/pubgo/lava/example/srv/perm/org"
	"github.com/pubgo/lava/example/srv/perm/perm"
	"github.com/pubgo/lava/example/srv/perm/role"

	_ "github.com/pubgo/lava/example/migrates"
)

var name = "perm"

func NewSrv() service.Service {
	srv := lava.NewSrv(name, "rbac grpc service")
	srv.Dix(menu.New)
	srv.Dix(group.New)
	srv.Dix(org.New)
	srv.Dix(perm.New)
	srv.Dix(role.New)
	srv.Dix(permpb.RegisterMenuServiceServer)
	srv.Dix(permpb.RegisterGroupServiceServer)
	srv.Dix(permpb.RegisterOrgServiceServer)
	srv.Dix(permpb.RegisterPermServiceServer)
	srv.Dix(permpb.RegisterRoleServiceServer)
	return srv
}
