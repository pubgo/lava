package restc

import (
	"context"

	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/types"
)

func doFunc(c *clientImpl) types.MiddleNext {
	var r = retry.New(retry.WithMaxRetries(c.cfg.RetryCount, c.cfg.backoff))
	return func(ctx context.Context, req types.Request, callback func(rsp types.Response) error) error {
		var req1 = req.(*Request).req.WithContext(ctx)
		return r.Do(func(i int) error {
			resp, err := c.client.Do(req1)
			if err != nil {
				return err
			}
			return callback(&Response{resp: resp})
		})
	}
}
