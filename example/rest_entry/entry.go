package rest_entry

import (
	"context"
	_ "expvar"
	"go.uber.org/zap"
	"net/http"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/types"
)

var name = "test-http"

func GetEntry() entry.Entry {
	ent := lug.NewRest(name)
	ent.Description("entry http test")

	ent.Middleware(func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			zap.L().Info("test http entry")
			return next(ctx, req, resp)
		}
	})

	ent.BeforeStart(func() {
		go http.ListenAndServe(":8083", nil)
	})

	ent.Register(&Service{})
	return ent
}
