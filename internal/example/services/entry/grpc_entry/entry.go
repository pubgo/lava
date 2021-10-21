package grpc_entry

import (
	"context"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/grpcEntry"
	"github.com/pubgo/lava/healthy"
	"github.com/pubgo/lava/internal/example/services/entry/grpc_entry/handler"
	"github.com/pubgo/lava/plugins/logger"
	"github.com/pubgo/lava/types"
)

var name = "test-grpc"

func GetEntry() entry.Entry {
	ent := grpcEntry.New(name)
	ent.Description("entry grpc test")
	ent.Register(handler.NewTestAPIHandler())
	ent.Middleware(func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			var log = logger.GetLog(ctx)
			log.Info("test grpc entry")
			return next(ctx, req, resp)
		}
	})

	// 健康检查
	healthy.Register(name, func(ctx context.Context) error {
		return nil
	})

	return ent
}
