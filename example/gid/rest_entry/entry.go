package rest_entry

import (
	"context"
	"go.uber.org/zap"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/restEntry"
	"github.com/pubgo/lug/types"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := restEntry.New(name)
	ent.Description("entry http test")
	ent.Middleware(func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			zap.L().Info("test http entry")
			return next(ctx, req, resp)
		}
	})
	ent.Register(&Service{})
	return ent
}
