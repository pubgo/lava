package grpc_entry

import (
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

type entryServerWrapper struct {
	*grpc.Server
	handlers []func()
}

func (t *entryServerWrapper) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	defer xerror.RespExit()

	t.Server.RegisterService(sd, handler)
}
