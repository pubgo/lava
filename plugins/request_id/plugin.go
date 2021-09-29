package request_id

import (
	"context"
	"fmt"

	"github.com/segmentio/ksuid"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
)

const Name = "request_id"

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
				var header = req.Header()
				var resID = header.Get(consts.HeaderXRequestID)
				if resID == "" {
					resID = ksuid.New().String()
					req.Header().Set(consts.HeaderXRequestID, resID)
				}

				defer func() {
					switch err := recover().(type) {
					case nil:
					case error:
						gErr = err
					default:
						gErr = fmt.Errorf("%#v\n", err)
					}
				}()

				return next(ctxFromReqId(ctx, resID), req, resp)
			}
		},
	})
}
