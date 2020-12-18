package golug_timeout

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_entry/golug_grpc"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var defaultTimeOut = time.Second

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {
			ent.UnWrap(func(entry golug_grpc.Entry) {
			})
		},
	}))
}

func TimeoutUnaryServerInterceptor(t time.Duration) grpc.UnaryServerInterceptor {
	if t := os.Getenv("GRPC_UNARY_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeOut = time.Duration(s) * time.Second
		}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := ctx.Deadline(); !ok { //if ok is true, it is set by header grpc-timeout from client
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
			defer cancel()
		}

		// create a done channel to tell the request it's done
		done := make(chan struct{})

		// here you put the actual work needed for the request
		// and then send the doneChan with the status and body
		// to finish the request by writing the response
		var res interface{}
		var err error

		go func() {
			defer func() {
				if c := recover(); c != nil {
					log.Errorf("response request panic: %v", c)
					err = status.Errorf(codes.Internal, "response request panic: %v", c)
				}
				close(done)
			}()
			res, err = handler(ctx, req)
		}()

		// non-blocking select on two channels see if the request
		// times out or finishes
		select {

		// if the context is done it timed out or was canceled
		// so don't return anything
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "handler timeout")

		// if the request finished then finish the request by
		// writing the response
		case <-done:
			return res, err
		}
	}
}

func TimeoutStreamServerInterceptor(defaultTimeOut time.Duration) grpc.StreamServerInterceptor {
	if t := os.Getenv("GRPC_STREAM_TIMEOUT"); t != "" {
		if s, err := strconv.Atoi(t); err == nil && s > 0 {
			defaultTimeOut = time.Duration(s) * time.Second
		}
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ctx := stream.Context()
		if _, ok := ctx.Deadline(); !ok { //if ok is true, it is set by header grpc-timeout from client
			if defaultTimeOut == 0 {
				return handler(srv, stream)
			}
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeOut)
			defer cancel()
		}
		// create a done channel to tell the request it's done
		done := make(chan struct{})

		go func() {
			defer func() {
				if c := recover(); c != nil {
					log.Errorf("response request panic: %v", c)
					err = status.Errorf(codes.Internal, "response request panic: %v", c)
				}
				close(done)
			}()
			err = handler(srv, stream)
		}()

		// non-blocking select on two channels see if the request
		// times out or finishes
		select {

		// if the context is done it timed out or was canceled
		// so don't return anything
		case <-ctx.Done():
			return status.Errorf(codes.DeadlineExceeded, "handler timeout")

		// if the request finished then finish the request by
		// writing the response
		case <-done:
			return err
		}
	}
}
