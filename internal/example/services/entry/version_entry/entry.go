package version_entry

import (
	"context"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/ginEntry"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/types"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := ginEntry.New(name)
	ent.Description("version http test")
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
