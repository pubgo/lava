package timeout

import (
	"context"
	"os"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

const Name = "timeout"

func init() {
	plugin.Middleware(Name, func(next types.MiddleNext) types.MiddleNext {
		var defaultTimeOut = consts.DefaultTimeout
		if t := os.Getenv("LAVA-TIMEOUT"); t != "" {
			var dur, err = time.ParseDuration(t)
			if dur != 0 && err == nil {
				defaultTimeOut = dur
			}
		}

		return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
			if t := req.Header().Get("LAVA-REQUEST-TIMEOUT"); t != "" {
				var dur, err = time.ParseDuration(t)
				if dur != 0 && err == nil {
					defaultTimeOut = dur
				}
			}

			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
				defer cancel()
			}

			done := make(chan struct{})
			var err error

			go func() {
				defer func() {
					close(done)
					if c := recover(); c != nil {
						err = status.Errorf(codes.Internal, "response request panic: %v", c)
					}
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
