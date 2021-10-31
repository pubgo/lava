package recovery

import (
	"context"
	"fmt"

	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const name = "recovery"

func init() {
	plugin.Middleware(name, func(next types.MiddleNext) types.MiddleNext {
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
			defer func() {
				switch err := recover().(type) {
				case nil:
				case error:
					gErr = err
				default:
					gErr = fmt.Errorf("%#v", err)
				}
			}()

			return next(ctx, req, resp)
		}
	})
}
