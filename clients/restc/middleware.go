package restc

import (
	"context"

	"github.com/pubgo/lava/pkg/retry"
	"github.com/pubgo/lava/service/service_type"
)

func doFunc(c *clientImpl) service_type.HandlerFunc {
	var r = retry.New(retry.WithMaxRetries(c.cfg.RetryCount, c.cfg.backoff))
	return func(ctx context.Context, req service_type.Request, callback func(rsp service_type.Response) error) error {
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
