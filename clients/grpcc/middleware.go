package grpcc

import (
	"context"
	"strings"
	"time"

	"github.com/pubgo/funk/errors"

	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/strutil"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/grpcutil"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/rs/xid"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func md2Head(md metadata.MD, header interface{ Add(key, value string) }) {
	for k, v := range md {
		for i := range v {
			header.Add(k, v[i])
		}
	}
}

func head2md(header *lava.RequestHeader, md metadata.MD) {
	header.VisitAll(func(key, value []byte) {
		md.Append(convert.BtoS(key), convert.BtoS(value))
	})
}

func unaryInterceptor(middlewares []lava.Middleware) grpc.UnaryClientInterceptor {
	unaryWrapper := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		md := make(metadata.MD)
		head2md(req.Header(), md)
		ctx = metadata.NewOutgoingContext(ctx, md)
		reqCtx := req.(*request)
		header := make(metadata.MD)
		trailer := make(metadata.MD)
		reqCtx.opts = append(reqCtx.opts, grpc.Header(&header), grpc.Trailer(&trailer))

		if err := reqCtx.invoker(ctx, reqCtx.method, reqCtx.req, reqCtx.reply, reqCtx.cc, reqCtx.opts...); err != nil {
			return nil, err
		}

		rsp := &response{resp: reqCtx.resp, header: new(lava.ResponseHeader)}
		md2Head(header, rsp.header)
		md2Head(trailer, rsp.header)
		return rsp, nil
	}

	unaryWrapper = lava.Chain(middlewares...).Middleware(unaryWrapper)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

		// get content type
		ct := strutil.FirstFnNotEmpty(func() string {
			return grpcutil.HeaderGet(md, "content-type")
		}, func() string {
			return grpcutil.HeaderGet(md, "x-content-type")
		}, func() string {
			return grpcc_config.DefaultContentType
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

		header := &fasthttp.RequestHeader{}
		md2Head(md, header)

		rpcReq := &request{
			ct:      ct,
			header:  header,
			service: serviceFromMethod(method),
			opts:    opts,
			method:  method,
			req:     req,
			cc:      cc,
			invoker: invoker,
			reply:   reply,
		}

		reqId := strutil.FirstFnNotEmpty(
			func() string { return lava.GetReqID(ctx) },
			func() string { return string(rpcReq.Header().Peek(httputil.HeaderXRequestID)) },
			func() string { return xid.New().String() },
		)
		rpcReq.Header().Set(httputil.HeaderXRequestID, reqId)
		ctx = lava.CreateCtxWithReqID(ctx, reqId)

		_, err = unaryWrapper(ctx, rpcReq)
		return errors.WrapCaller(err)
	}
}

func streamInterceptor(middlewares []lava.Middleware) grpc.StreamClientInterceptor {
	wrapperStream := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		reqCtx := req.(*request)
		md := make(metadata.MD)
		head2md(req.Header(), md)

		ctx = metadata.NewOutgoingContext(ctx, md)
		stream, err := reqCtx.streamer(ctx, reqCtx.desc, reqCtx.cc, reqCtx.method, reqCtx.opts...)
		if err != nil {
			return nil, err
		}

		return &response{header: new(lava.ResponseHeader), stream: stream}, nil
	}

	wrapperStream = lava.Chain(middlewares...).Middleware(wrapperStream)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (resp grpc.ClientStream, err error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = make(metadata.MD)
		}

		// get content type
		ct := strutil.FirstFnNotEmpty(func() string {
			return grpcutil.HeaderGet(md, "content-type")
		}, func() string {
			return grpcutil.HeaderGet(md, "x-content-type")
		}, func() string {
			return grpcc_config.DefaultContentType
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

		header := &fasthttp.RequestHeader{}
		md2Head(md, header)

		reqId := strutil.FirstFnNotEmpty(
			func() string { return lava.GetReqID(ctx) },
			func() string { return string(header.Peek(httputil.HeaderXRequestID)) },
			func() string { return xid.New().String() },
		)
		header.Set(httputil.HeaderXRequestID, reqId)
		ctx = lava.CreateCtxWithReqID(ctx, reqId)

		rsp, err := wrapperStream(ctx,
			&request{
				ct:       ct,
				service:  serviceFromMethod(method),
				header:   header,
				opts:     opts,
				desc:     desc,
				cc:       cc,
				method:   method,
				streamer: streamer,
			},
		)

		return rsp.(*response).stream, err
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
