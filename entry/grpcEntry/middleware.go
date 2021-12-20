package grpcEntry

import (
	"context"
	"strings"
	"time"

	grpcMiddle "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/encoding"
	"github.com/pubgo/lava/types"
)

func (g *grpcEntry) handlerUnaryMiddle(middlewares []types.Middleware) grpc.UnaryServerInterceptor {
	unaryWrapper := func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		if len(req.Header()) > 0 {
			_ = grpc.SetHeader(ctx, req.Header())
		}

		dt, err := req.(*rpcRequest).handler(ctx, req.Payload())
		if err != nil {
			return err
		}
		return rsp(&rpcResponse{dt: dt, header: req.Header()})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		unaryWrapper = middlewares[i](unaryWrapper)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var md, ok = metadata.FromIncomingContext(ctx)
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
		if p := getPeerIP(md, ctx); p != "" {
			md.Set("remote-ip", p)
		}

		if p := getPeerName(md); p != "" {
			md.Set("remote-name", p)
		}

		// timeout for server deadline
		to := md.Get("timeout")
		delete(md, "timeout")

		// set the timeout if we have it
		if len(to) != 0 && to[0] != "" {
			if dur, err := time.ParseDuration(to[0]); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, dur)
				_ = cancel
			}
		}

		var cdc = encoding.GetCdc(ct)
		xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)

		err = unaryWrapper(ctx,
			&rpcRequest{
				service:     serviceFromMethod(info.FullMethod),
				method:      info.FullMethod,
				handler:     handler,
				contentType: ct,
				cdc:         cdc,
				payload:     req,
				header:      md,
			},
			func(rsp types.Response) error { resp = rsp.Payload(); return nil },
		)

		return
	}
}

func (g *grpcEntry) handlerStreamMiddle(middlewares []types.Middleware) grpc.StreamServerInterceptor {
	streamWrapper := func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		ctx = metadata.NewIncomingContext(ctx, req.Header())
		var reqCtx = req.(*rpcRequest)
		err := reqCtx.handlerStream(reqCtx.srv, &grpcMiddle.WrappedServerStream{WrappedContext: ctx, ServerStream: reqCtx.stream})
		if err != nil {
			return err
		}
		return rsp(&rpcResponse{stream: reqCtx.stream, header: req.Header()})
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		streamWrapper = middlewares[i](streamWrapper)
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var ctx = stream.Context()
		var md, ok = metadata.FromIncomingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

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
				defer cancel()
			}
		}

		var cdc = encoding.GetCdc(ct)
		xerror.Assert(cdc == nil, "contentType(%s) codec not found", ct)

		return streamWrapper(ctx,
			&rpcRequest{
				stream:        stream,
				srv:           srv,
				handlerStream: handler,
				header:        md,
				method:        info.FullMethod,
				service:       serviceFromMethod(info.FullMethod),
				contentType:   ct,
				cdc:           cdc,
			},
			func(_ types.Response) error { return nil },
		)
	}
}

// serviceFromMethod returns the service
// /service.Foo/Bar => service
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	if m[0] != '/' {
		return m
	}

	parts := strings.Split(m, "/")
	if len(parts) < 3 {
		return m
	}

	parts = strings.Split(parts[1], ".")
	return strings.Join(parts[:len(parts)-1], ".")
}
