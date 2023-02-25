package grpcs

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	grpcMiddle "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/recovery"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/grpcutil"
)

func parsePath(path string) (string, string, string) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", ""
	}
	method := parts[len(parts)-1]
	pkgService := parts[len(parts)-2]
	prefix := strings.Join(parts[0:len(parts)-2], "/")
	return prefix, pkgService, method
}

func handlerTwMiddle(middlewares map[string][]lava.Middleware, handle http.Handler) func(fbCtx *fiber.Ctx) error {
	var handler = func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var reqCtx = req.(*httpRequest)
		var err = adaptor.HTTPHandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			handle.ServeHTTP(writer, request.WithContext(ctx))
		})(reqCtx.ctx)
		if err != nil {
			return nil, err
		}
		return &httpResponse{ctx: reqCtx.ctx}, nil
	}

	return func(fbCtx *fiber.Ctx) error {
		var prefix, srv, _ = parsePath(fbCtx.OriginalURL())
		if prefix == "" || srv == "" {
			return fmt.Errorf("invalid path, path=%q", fbCtx.OriginalURL())
		}

		_, err := lava.Chain(middlewares[srv]...)(handler)(fbCtx.Context(), &httpRequest{ctx: fbCtx})
		if err != nil {
			return err
		}

		return nil
	}
}

func handlerUnaryMiddle(middlewares map[string][]lava.Middleware) grpc.UnaryServerInterceptor {
	unaryWrapper := func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
		// 错误和panic处理
		defer recovery.Err(&gErr, func(err *errors.Event) {
			err.Str("stack", string(debug.Stack()))
			err.Str("operation", req.Operation())
		})

		var md = make(metadata.MD)
		req.Header().VisitAll(func(key, value []byte) {
			md.Append(convert.BtoS(key), convert.BtoS(value))
		})

		ctx = metadata.NewIncomingContext(ctx, md)
		dt, err := req.(*rpcRequest).handler(ctx, req.Payload())
		if err != nil {
			return nil, err
		}

		return &rpcResponse{header: new(fasthttp.ResponseHeader), dt: dt}, nil
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
		if p := grpcutil.ClientIP(md); p != "" {
			md.Set("remote-ip", p)
		}

		if p := grpcutil.ClientName(md); p != "" {
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

		var srvName = serviceFromMethod(info.FullMethod)
		var rpcReq = &rpcRequest{
			service:     srvName,
			method:      info.FullMethod,
			url:         url,
			handler:     handler,
			contentType: ct,
			payload:     req,
			header:      header,
		}

		var rsp, err = lava.Chain(middlewares[srvName]...)(unaryWrapper)(ctx, rpcReq)
		if err != nil {
			return nil, err
		}

		var h = rsp.Header()
		md = make(metadata.MD, h.Len())
		h.VisitAll(func(key, value []byte) {
			md.Append(convert.BtoS(key), convert.BtoS(value))
		})

		return rsp.(*rpcResponse).dt, grpc.SetTrailer(ctx, md)
	}
}

func handlerStreamMiddle(middlewares map[string][]lava.Middleware) grpc.StreamServerInterceptor {
	streamWrapper := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var md = make(metadata.MD)
		req.Header().VisitAll(func(key, value []byte) {
			md.Append(convert.BtoS(key), convert.BtoS(value))
		})

		ctx = metadata.NewIncomingContext(ctx, md)
		var reqCtx = req.(*rpcRequest)
		if err := reqCtx.handlerStream(reqCtx.srv, &grpcMiddle.WrappedServerStream{WrappedContext: ctx, ServerStream: reqCtx.stream}); err != nil {
			return nil, err
		}

		return &rpcResponse{stream: reqCtx.stream, header: new(lava.ResponseHeader)}, nil
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

		srvName := serviceFromMethod(info.FullMethod)
		var rpcReq = &rpcRequest{
			stream:        stream,
			srv:           srv,
			handlerStream: handler,
			header:        header,
			method:        info.FullMethod,
			service:       srvName,
			contentType:   ct,
		}

		var rsp, err = lava.Chain(middlewares[srvName]...)(streamWrapper)(ctx, rpcReq)
		if err != nil {
			return err
		}

		var h = rsp.Header()
		md = make(metadata.MD)
		h.VisitAll(func(key, value []byte) {
			md.Append(convert.BtoS(key), convert.BtoS(value))
		})
		return grpc.SetTrailer(ctx, md)
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
