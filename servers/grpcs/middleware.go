package grpcs

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	grpcMiddle "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/convert"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/errors/errcode"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/proto/errorpb"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/v2/core/lavacontexts"
	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/grpcutil"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
)

func handlerUnaryMiddle(middlewares map[string][]lava.Middleware) grpc.UnaryServerInterceptor {
	unaryWrapper := func(ctx context.Context, req lava.Request) (rsp lava.Response, gErr error) {
		dt, err := req.(*rpcRequest).handler(ctx, req.Payload())
		if err != nil {
			return nil, err
		}

		return &rpcResponse{header: req.(*rpcRequest).rspHeader, dt: dt}, nil
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		reqMetadata := getIncomingMetadata(ctx)

		// get content type
		ct := defaultContentType
		if c := reqMetadata.Get("x-content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		if c := reqMetadata.Get("content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}

		delete(reqMetadata, "x-content-type")

		clientInfo := new(lavapbv1.ServiceInfo)

		// get peer from context
		if p := grpcutil.ClientIP(reqMetadata); p != "" {
			clientInfo.Ip = p
		}

		if p := grpcutil.ClientName(reqMetadata); p != "" {
			clientInfo.Name = p
		}

		// get peer from context
		if p, ok := peer.FromContext(ctx); ok {
			reqMetadata.Set("remote", p.Addr.String())
		}

		// timeout for server deadline
		to := reqMetadata.Get("timeout")
		delete(reqMetadata, "timeout")

		// set the timeout if we have it
		if len(to) != 0 && to[0] != "" {
			if dur, err := time.ParseDuration(to[0]); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, dur)
				defer cancel()
			}
		}

		// 从gateway获取url
		url := info.FullMethod
		if _url, ok := reqMetadata["url"]; ok {
			url = _url[0]
		}

		reqHeader := &fasthttp.RequestHeader{}
		for k, v := range reqMetadata {
			for i := range v {
				reqHeader.Add(k, v[i])
			}
		}

		srvName := serviceFromMethod(info.FullMethod)
		rpcReq := &rpcRequest{
			service:     srvName,
			method:      info.FullMethod,
			url:         url,
			handler:     handler,
			contentType: ct,
			payload:     req,
			header:      reqHeader,
			rspHeader:   new(fasthttp.ResponseHeader),
		}

		reqId := strutil.FirstFnNotEmpty(
			func() string { return lavacontexts.GetReqID(ctx) },
			func() string { return string(rpcReq.Header().Peek(httputil.HeaderXRequestID)) },
			func() string { return xid.New().String() },
		)
		ctx = lavacontexts.CreateCtxWithReqID(ctx, reqId)

		reqHeader.Set(httputil.HeaderXRequestID, reqId)
		reqHeader.Set(httputil.HeaderXRequestVersion, version.Version())

		defer func() {
			reqMetadata = make(metadata.MD)
			reqMetadata.Set(httputil.HeaderXRequestID, reqId)
			reqMetadata.Set(httputil.HeaderXRequestVersion, version.Version())
			reqMetadata.Set(httputil.HeaderXRequestOperation, info.FullMethod)
			for key, value := range rpcReq.rspHeader.All() {
				reqMetadata.Set(convert.BtoS(key), convert.BtoS(value))
			}

			if err := grpc.SendHeader(ctx, reqMetadata); err != nil {
				log.Err(err, ctx).
					Str("grpc-method", info.FullMethod).
					Msg("grpc send trailer header failed")
			}
		}()

		ctx = lavacontexts.CreateReqHeader(ctx, reqHeader)
		ctx = lavacontexts.CreateRspHeader(ctx, rpcReq.rspHeader)
		rsp, err := lava.Chain(middlewares[srvName]...).Middleware(unaryWrapper)(ctx, rpcReq)
		if err != nil {
			pb := errcode.ParseError(err)
			pb.Details = append(pb.Details, errcode.MustTagsToAny(errors.Tags{"reqHeader": string(rpcReq.Header().Header())})...)

			if pb.Message == "" {
				pb.Message = err.Error()
			}

			if pb.Code == 0 {
				pb.StatusCode = errorpb.Code_Internal
				pb.Code = int32(errcode.GrpcCodeToHTTP(codes.Code(errorpb.Code_Internal)))
			}

			return nil, errcode.ConvertErr2Status(pb).Err()
		}

		return rsp.(*rpcResponse).dt, nil
	}
}

func handlerStreamMiddle(middlewares map[string][]lava.Middleware) grpc.StreamServerInterceptor {
	streamWrapper := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*rpcRequest)
		wrap := &grpcMiddle.WrappedServerStream{WrappedContext: ctx, ServerStream: reqCtx.stream}
		if err := reqCtx.handlerStream(reqCtx.srv, wrap); err != nil {
			return nil, err
		}

		return &rpcResponse{stream: reqCtx.stream, header: new(lava.ResponseHeader)}, nil
	}

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		md := getIncomingMetadata(stream.Context())

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

		header := new(fasthttp.RequestHeader)
		for k, v := range md {
			for i := range v {
				header.Add(k, v[i])
			}
		}

		srvName := serviceFromMethod(info.FullMethod)
		rpcReq := &rpcRequest{
			stream:        stream,
			srv:           srv,
			handlerStream: handler,
			header:        header,
			method:        info.FullMethod,
			service:       srvName,
			contentType:   ct,
		}

		reqId := strutil.FirstFnNotEmpty(
			func() string { return lavacontexts.GetReqID(ctx) },
			func() string { return string(rpcReq.Header().Peek(httputil.HeaderXRequestID)) },
			func() string { return xid.New().String() },
		)
		rpcReq.Header().Set(httputil.HeaderXRequestID, reqId)

		ctx = lavacontexts.CreateCtxWithReqID(ctx, reqId)
		ctx = lavacontexts.CreateReqHeader(ctx, header)
		ctx = lavacontexts.CreateRspHeader(ctx, rpcReq.rspHeader)
		rsp, err := lava.Chain(middlewares[srvName]...).Middleware(streamWrapper)(ctx, rpcReq)
		if err != nil {
			pb := errcode.ParseError(err)
			if pb.Message == "" {
				pb.Message = err.Error()
			}

			if pb.Code == 0 {
				pb.StatusCode = errorpb.Code_Internal
				pb.Code = int32(errcode.GrpcCodeToHTTP(codes.Code(errorpb.Code_Internal)))
			}

			return errcode.ConvertErr2Status(pb).Err()
		}

		h := rsp.Header()
		md = make(metadata.MD)
		for key, value := range h.All() {
			md.Append(convert.BtoS(key), convert.BtoS(value))
		}
		return grpc.SendHeader(ctx, md)
	}
}

func handlerHttpMiddle(middlewares []lava.Middleware) func(fbCtx *fiber.Ctx) error {
	h := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*httpRequest).ctx
		reqCtx.SetUserContext(ctx)
		return &httpResponse{ctx: reqCtx}, reqCtx.Next()
	}

	h = lava.Chain(middlewares...).Middleware(h)
	return func(ctx *fiber.Ctx) error {
		_, err := h(ctx.Context(), &httpRequest{ctx: ctx})
		return err
	}
}
