package grpcc

import (
	"context"
	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/pkg/lavax"
	"time"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/types"
)

func unaryInterceptor(middlewares []types.Middleware) grpc.UnaryClientInterceptor {
	var unaryWrapper = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		var reqCtx = req.(*request)
		ctx = metadata.NewOutgoingContext(ctx, reqCtx.Header())
		if err := reqCtx.invoker(ctx, reqCtx.method, reqCtx.req, reqCtx.reply, reqCtx.cc); err != nil {
			return err
		}
		return rsp(&response{req: reqCtx, resp: reqCtx.reply})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		unaryWrapper = middlewares[i](unaryWrapper)
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		var md, ok = metadata.FromOutgoingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

		// get content type
		ct := lavax.FirstNotEmpty(func() string {
			return types.HeaderGet(md, "content-type")
		}, func() string {
			return types.HeaderGet(md, "x-content-type")
		}, func() string {
			return defaultContentType
		})

		delete(md, "x-content-type")

		// get peer from context
		if p, _ok := peer.FromContext(ctx); _ok {
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
				defer cancel()
			}
		}

		var cdc = encoding.GetCdc(ct)
		xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)

		return unaryWrapper(ctx,
			&request{
				ct:      ct,
				cdc:     cdc,
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
		ct := lavax.FirstNotEmpty(func() string {
			return types.HeaderGet(md, "content-type")
		}, func() string {
			return types.HeaderGet(md, "x-content-type")
		}, func() string {
			return defaultContentType
		})

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
				defer cancel()
			}
		}

		var cdc = encoding.GetCdc(ct)
		xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)

		return nil, wrapperStream(ctx,
			&request{
				ct:       ct,
				cdc:      cdc,
				service:  serviceFromMethod(method),
				header:   md,
				opts:     opts,
				desc:     desc,
				cc:       cc,
				method:   method,
				streamer: streamer,
			},
			func(rsp types.Response) error { resp = rsp.(*response).stream; return nil },
		)
	}
}
