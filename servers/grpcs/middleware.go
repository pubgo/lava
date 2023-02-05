package grpcs

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	grpcMiddle "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/pkg/grpcutil"
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/service"
)

func handlerHttpMiddle(middlewares []service.Middleware) func(fbCtx *fiber.Ctx) error {
	var handler = func(ctx context.Context, req service.Request, rsp service.Response) error {
		var reqCtx = req.(*httpRequest)
		reqCtx.ctx.SetUserContext(ctx)
		return reqCtx.ctx.Next()
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return func(fbCtx *fiber.Ctx) error {
		return handler(fbCtx.Context(), &httpRequest{ctx: fbCtx}, &httpResponse{ctx: fbCtx})
	}
}

func handlerUnaryMiddle(middlewares []service.Middleware) grpc.UnaryServerInterceptor {
	unaryWrapper := func(ctx context.Context, req service.Request, rsp service.Response) (gErr error) {
		// 错误和panic处理
		defer recovery.Err(&gErr, func(err *errors.Event) {
			err.Str("stack", string(debug.Stack()))
			err.Str("operation", req.Operation())
		})

		var md = make(metadata.MD)
		req.Header().VisitAll(func(key, value []byte) {
			md.Append(utils.BtoS(key), utils.BtoS(value))
		})

		ctx = metadata.NewIncomingContext(ctx, md)
		dt, err := req.(*rpcRequest).handler(ctx, req.Payload())
		if err != nil {
			return err
		}

		rsp.(*rpcResponse).dt = dt
		var h = rsp.(*rpcResponse).Header()
		md = make(metadata.MD)
		h.VisitAll(func(key, value []byte) {
			md.Append(utils.BtoS(key), utils.BtoS(value))
		})
		return grpc.SetTrailer(ctx, md)
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		unaryWrapper = middlewares[i](unaryWrapper)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		if p := grpcutil.GetClientIP(md); p != "" {
			md.Set("remote-ip", p)
		}

		if p := grpcutil.GetClientName(md); p != "" {
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

		// 从gateway获取url
		var url = info.FullMethod
		if _url, ok := md["url"]; ok {
			url = _url[0]
		}

		var header = &fasthttp.RequestHeader{}
		for k, v := range md {
			for i := range v {
				header.Add(k, v[i])
			}
		}

		var rpcReq = &rpcRequest{
			service:     serviceFromMethod(info.FullMethod),
			method:      info.FullMethod,
			url:         url,
			handler:     handler,
			contentType: ct,
			payload:     req,
			header:      header,
		}

		var rpcResp = &rpcResponse{header: new(fasthttp.ResponseHeader)}
		return rpcResp.dt, unaryWrapper(ctx, rpcReq, rpcResp)
	}
}

func handlerStreamMiddle(middlewares []service.Middleware) grpc.StreamServerInterceptor {
	streamWrapper := func(ctx context.Context, req service.Request, rsp service.Response) error {
		var md = make(metadata.MD)
		req.Header().VisitAll(func(key, value []byte) {
			md.Append(utils.BtoS(key), utils.BtoS(value))
		})

		ctx = metadata.NewIncomingContext(ctx, md)
		var reqCtx = req.(*rpcRequest)
		if err := reqCtx.handlerStream(reqCtx.srv, &grpcMiddle.WrappedServerStream{WrappedContext: ctx, ServerStream: reqCtx.stream}); err != nil {
			return err
		}

		var h = rsp.(*rpcResponse).Header()
		md = make(metadata.MD)
		h.VisitAll(func(key, value []byte) {
			md.Append(utils.BtoS(key), utils.BtoS(value))
		})
		return grpc.SetTrailer(ctx, md)
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

		var header = &fasthttp.RequestHeader{}
		for k, v := range md {
			for i := range v {
				header.Add(k, v[i])
			}
		}

		var rpcReq = &rpcRequest{
			stream:        stream,
			srv:           srv,
			handlerStream: handler,
			header:        header,
			method:        info.FullMethod,
			service:       serviceFromMethod(info.FullMethod),
			contentType:   ct,
		}
		return streamWrapper(ctx, rpcReq, &rpcResponse{stream: stream, header: new(service.ResponseHeader)})
	}
}

// serviceFromMethod returns the service
// /service.Foo/Bar => service.Foo
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	return strings.Split(strings.Trim(m, "/"), "/")[0]
}
