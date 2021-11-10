package grpcc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/types"
)

func unaryInterceptor(middlewares []types.Middleware) grpc.UnaryClientInterceptor {
	var wrapperUnary = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*request)
		ctx = metadata.NewOutgoingContext(ctx, reqCtx.Header())
		if err := reqCtx.invoker(ctx, reqCtx.method, reqCtx.req, reqCtx.reply, reqCtx.cc); err != nil {
			return err
		}

		return rsp(&response{req: reqCtx, resp: reqCtx.reply})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapperUnary = middlewares[i](wrapperUnary)
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var md, ok = metadata.FromOutgoingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

		// get content type
		ct := defaultContentType
		if c := md.Get("x-content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		if c := md.Get("content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		delete(md, "x-content-type")

		// get peer from context
		if p, ok := peer.FromContext(ctx); ok {
			md.Set("remote", p.Addr.String())
		}

		// timeout for server deadline
		to := md.Get("timeout")
		delete(md, "timeout")

		// set the timeout if we have it
		if len(to) != 0 {
			if dur, err := time.ParseDuration(to[0]); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, dur)
				_ = cancel
			}
		}

		return wrapperUnary(
			ctx,
			&request{
				ct:      ct,
				header:  md,
				service: serviceFromMethod(method),
				opts:    opts,
				method:  method,
				req:     req,
				reply:   reply,
				cc:      cc,
				invoker: invoker,
			},
			func(_ types.Response) error { return nil },
		)
	}
}

func streamInterceptor(middlewares []types.Middleware) grpc.StreamClientInterceptor {
	wrapperStream := func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*request)
		ctx = metadata.NewOutgoingContext(ctx, reqCtx.Header())
		stream, err := reqCtx.streamer(ctx, reqCtx.desc, reqCtx.cc, reqCtx.method)
		if err != nil {
			return err
		}

		return rsp(&response{req: reqCtx, stream: stream})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapperStream = middlewares[i](wrapperStream)
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (resp grpc.ClientStream, err error) {
		var md, ok = metadata.FromOutgoingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

		// get content type
		ct := defaultContentType
		if c := md.Get("x-content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		if c := md.Get("content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		delete(md, "x-content-type")

		// get peer from context
		if p, ok := peer.FromContext(ctx); ok {
			md.Set("remote", p.Addr.String())
		}

		// timeout for server deadline
		to := md.Get("timeout")
		delete(md, "timeout")

		// set the timeout if we have it
		if len(to) != 0 {
			if dur, err := time.ParseDuration(to[0]); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, dur)
				_ = cancel
			}
		}

		return nil, wrapperStream(
			ctx,
			&request{
				ct:       ct,
				service:  serviceFromMethod(method),
				header:   md,
				opts:     opts,
				desc:     desc,
				cc:       cc,
				method:   method,
				streamer: streamer,
			},
			func(rsp types.Response) error {
				resp = rsp.(*response).stream
				return nil
			},
		)
	}
}