package rest_entry

import (
	"context"
	"github.com/pubgo/lava/plugins/logger"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/restEntry"
	"github.com/pubgo/lava/types"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := restEntry.New(name)
	ent.Description("entry http test")
	ent.Middleware(func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			var log = logger.GetLog(ctx)
			log.Info("test http entry")
			return next(ctx, req, resp)
		}
	})
	ent.Register(&Service{})
	return ent
}
