package grpc

import (
	"context"
	"strings"
	"time"

	"github.com/pubgo/lug/types"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func (g *grpcEntry) handlerUnaryMiddle(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var wrapper = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		dt, err := handler(ctx, req.Payload())
		if err != nil {
			return xerror.Wrap(err)
		}

		return xerror.Wrap(rsp(&rpcResponse{dt: dt, header: req.Header()}))
	}

	var middlewares = g.Options().Middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapper = middlewares[i](wrapper)
	}

	// get grpc metadata
	gmd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		gmd = metadata.MD{}
	}

	var md = types.Header(gmd)

	// get content type
	ct := defaultContentType
	if c := md.Get("x-content-type"); c != "" {
		ct = c
	}

	if c := md.Get("content-type"); c != "" {
		ct = c
	}

	md.Del("x-content-type")

	// get peer from context
	if p, ok := peer.FromContext(ctx); ok {
		md.Set("remote", p.Addr.String())
	}

	// create a client.Request
	request := &rpcRequest{
		service: ServiceFromMethod(info.FullMethod),
		method:  info.FullMethod,

		contentType: ct,
		cdc:         ct,
		payload:     req,
		header:      types.Header(gmd),
	}

	// timeout for server deadline
	to := md.Get("timeout")
	md.Del("timeout")

	// set the timeout if we have it
	if to != "" {
		if dur, err := time.ParseDuration(to); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, dur)
			_ = cancel
		}
	}

	return resp, wrapper(ctx, request, func(rsp types.Response) error { resp = rsp.Payload(); return nil })
}

func (g *grpcEntry) handlerStreamMiddle(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	var wrapper = func(ctx context.Context, req types.Request, rsp func(response types.Response) error) error {
		err := handler(ctx, stream)
		if err != nil {
			return xerror.Wrap(err)
		}

		return xerror.Wrap(rsp(&rpcResponse{stream: stream, header: req.Header()}))
	}

	var middlewares = g.Options().Middlewares
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapper = middlewares[i](wrapper)
	}

	var ctx = stream.Context()

	// get grpc metadata
	gmd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		gmd = metadata.MD{}
	}

	var md = types.Header(gmd)

	// get content type
	ct := defaultContentType
	if c := md.Get("x-content-type"); c != "" {
		ct = c
	}

	if c := md.Get("content-type"); c != "" {
		ct = c
	}

	md.Del("x-content-type")

	// get peer from context
	if p, ok := peer.FromContext(ctx); ok {
		md.Set("remote", p.Addr.String())
	}

	// create a client.Request
	request := &rpcRequest{
		header:      types.Header(gmd),
		method:      info.FullMethod,
		service:     ServiceFromMethod(info.FullMethod),
		contentType: ct,
		cdc:         ct,
		payload:     stream,
	}

	// timeout for server deadline
	to := md.Get("timeout")
	md.Del("timeout")

	// set the timeout if we have it
	if to != "" {
		if dur, err := time.ParseDuration(to); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, dur)
			_ = cancel
		}
	}

	return wrapper(ctx, request, func(_ types.Response) error { return nil })
}

// ServiceFromMethod returns the service
// /service.Foo/Bar => service
func ServiceFromMethod(m string) string {
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
