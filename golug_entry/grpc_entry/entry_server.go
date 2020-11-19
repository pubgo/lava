package grpc_entry

import (
	"reflect"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type entryServerWrapper struct {
	*grpc.Server
}

func (t *entryServerWrapper) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	defer xerror.RespExit()

	t.Server.RegisterService(sd, ss)

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		grpclog.Fatalf("grpc: Server.Register found the handler of type %v that does not satisfy %v", st, ht)
	}

}
