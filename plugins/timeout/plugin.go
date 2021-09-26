package timeout

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	plugin.Register(&plugin.Base{
		Name:         "timeout",
		OnMiddleware: Middleware,
	})
}

func Middleware(next types.MiddleNext) types.MiddleNext {
	var defaultTimeOut = time.Second
	if t := os.Getenv("GRPC_UNARY_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeOut = time.Duration(s) * time.Second
		}
	}

	return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
		if _, ok := ctx.Deadline(); !ok { //if ok is true, it is set by header grpc-timeout from client
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
			defer cancel()
		}

		done := make(chan struct{})
		var err error

		go func() {
			defer func() {
				if c := recover(); c != nil {
					err = status.Errorf(codes.Internal, "response request panic: %v", c)
				}
				close(done)
			}()
			err = next(ctx, req, resp)
		}()

		select {
		case <-ctx.Done():
			return status.Errorf(codes.DeadlineExceeded, "handler timeout")
		case <-done:
			return err
		default:
			return nil
		}
	}
}
