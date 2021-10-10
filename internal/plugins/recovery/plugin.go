package recovery

import (
	"context"
	"fmt"

	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
)

const name = "recovery"

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) (gErr error) {
				defer func() {
					switch err := recover().(type) {
					case nil:
					case error:
						gErr = err
					default:
						gErr = fmt.Errorf("%#v\n", err)
					}
				}()

				return next(ctx, req, resp)
			}
		},
	})
}
