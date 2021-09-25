package request_id

import (
	"context"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"

	"github.com/segmentio/ksuid"
)

const Name = "request_id"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnMiddleware: func() types.Middleware {
			return func(next types.MiddleNext) types.MiddleNext {
				return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
					var header = req.Header()
					var resID = header.Get(consts.HeaderXRequestID)
					if resID == "" {
						resID = ksuid.New().String()
						req.Header().Set(consts.HeaderXRequestID, resID)
					}
					return next(ctxFromReqId(ctx, resID), req, resp)
				}
			}
		},
	})
}
