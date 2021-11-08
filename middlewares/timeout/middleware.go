package timeout

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/httpx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "timeout"

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		var defaultTimeout = consts.DefaultTimeout
		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			// 过滤 websocket 请求
			if httpx.IsWebsocket(http.Header(req.Header())) {
				return nil
			}

			if t := types.HeaderGet(req.Header(), "X-REQUEST-TIMEOUT"); t != "" {
				var dur, err = time.ParseDuration(t)
				if dur != 0 && err == nil {
					defaultTimeout = dur
				}
			}

			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
				defer cancel()
			}

			var err error
			var done = make(chan struct{})
			go func() {
				defer func() {
					switch c := recover().(type) {
					case nil:
					case error:
						err = c
					default:
						err = status.Errorf(codes.Internal, "service=>%s, endpoint=>%s, msg=>%v", req.Service(), req.Endpoint(), err)
					}
					close(done)
				}()

				err = next(ctx, req, resp)
			}()

			select {
			case <-ctx.Done():
				return status.Error(codes.DeadlineExceeded, req.Endpoint())
			case <-done:
				return err
			}
		}
	})
}
