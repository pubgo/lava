package grpc_entry

import (
	"context"
	"net/http"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/example/entry/grpc_entry/handler"
	"github.com/pubgo/lava/example/protopb/proto/hello"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/plugins/healthy"
	"github.com/pubgo/lava/types"
)

var name = "test-grpc"

func GetEntry() entry.Entry {
	ent := grpcEntry.New(name)
	ent.Description("entry grpc test")
	ent.Middleware(func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			var log = logging.GetLog(ctx)
			log.Info("test grpc entry")
			return next(ctx, req, resp)
		}
	})

	hello.RegisterTestApiSrvServer(ent, handler.NewTestAPIHandler())

	// 健康检查
	healthy.Register(name, func(req *http.Request) error {
		return nil
	})

	return ent
}
