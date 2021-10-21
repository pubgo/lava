package requestID

import (
	"context"
	"fmt"

	"github.com/segmentio/ksuid"

	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "x-request-id"

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
			var header = req.Header()
			var resID = header.Get(httpx.HeaderXRequestID)
			if resID == "" {
				resID = ksuid.New().String()
				req.Header().Set(httpx.HeaderXRequestID, resID)
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

			return next(WithReqID(ctx, resID), req, resp)
		}
	})
}
