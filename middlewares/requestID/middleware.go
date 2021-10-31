package requestID

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "x-request-id"

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
			var header = req.Header()
			var resID = header.Get(Name)
			var resIDStr string
			if len(resID) == 0 || resID[0] == "" {
				resIDStr = GetWith(ctx)
				req.Header().Set(Name, resIDStr)
			}

			defer func() {
				switch err := recover().(type) {
				case nil:
				case error:
					gErr = err
				default:
					gErr = fmt.Errorf("%#v", err)
				}
			}()

			return next(WithReqID(ctx, resIDStr), req, resp)
		}
	})
}
