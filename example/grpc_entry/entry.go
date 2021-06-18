package grpc_entry

import (
	"context"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/example/grpc_entry/handler"
	"google.golang.org/grpc"
)

var name = "test-grpc"

func GetEntry() entry.Entry {
	ent := lug.NewGrpc(name)
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	ent.Init()
	ent.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//q.Q(info)
		return handler(ctx, req)
	})
	return ent
}
