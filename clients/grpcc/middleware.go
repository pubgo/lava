package grpcc

import (
	"context"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	middleware2 "github.com/pubgo/lava/core/middleware"
	"github.com/pubgo/lava/internal/pkg/grpcutil"
	utils2 "github.com/pubgo/lava/internal/pkg/utils"
)

func md2Head(md metadata.MD, header interface{ Add(key, value string) }) {
	for k, v := range md {
		for i := range v {
			header.Add(k, v[i])
		}
	}
}

func head2md(header interface {
	VisitAll(f func(key, value []byte))
}, md metadata.MD) {
	header.VisitAll(func(key, value []byte) {
		md.Append(utils2.BtoS(key), utils2.BtoS(value))
	})
}

func unaryInterceptor(middlewares []middleware2.Middleware) grpc.UnaryClientInterceptor {
	var unaryWrapper = func(ctx context.Context, req middleware2.Request, rsp middleware2.Response) error {
		var md = make(metadata.MD)
		head2md(req.Header(), md)
		ctx = metadata.NewOutgoingContext(ctx, md)
		var reqCtx = req.(*request)
		var header = make(metadata.MD)
		var trailer = make(metadata.MD)
		reqCtx.opts = append(reqCtx.opts, grpc.Header(&header), grpc.Trailer(&trailer))

		if err := reqCtx.invoker(ctx, reqCtx.method, reqCtx.req, rsp.(*response).resp, reqCtx.cc, reqCtx.opts...); err != nil {
			return err
		}

		md2Head(header, rsp.(*response).header)
		md2Head(trailer, rsp.(*response).header)
		return nil
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
		ct := utils2.FirstFnNotEmpty(func() string {
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

		var header = &fasthttp.RequestHeader{}
		md2Head(md, header)

		return unaryWrapper(ctx,
			&request{
				ct:      ct,
				header:  header,
				service: serviceFromMethod(method),
				opts:    opts,
				method:  method,
				req:     req,
				cc:      cc,
				invoker: invoker,
			},
			&response{resp: reply, header: new(middleware2.ResponseHeader)},
		)
	}
}

func streamInterceptor(middlewares []middleware2.Middleware) grpc.StreamClientInterceptor {
	wrapperStream := func(ctx context.Context, req middleware2.Request, rsp middleware2.Response) error {
		var reqCtx = req.(*request)
		var md = make(metadata.MD)
		head2md(req.Header(), md)

		ctx = metadata.NewOutgoingContext(ctx, md)
		stream, err := reqCtx.streamer(ctx, reqCtx.desc, reqCtx.cc, reqCtx.method, reqCtx.opts...)
		if err != nil {
			return err
		}
		rsp.(*response).stream = stream
		return nil
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
		ct := utils2.FirstFnNotEmpty(func() string {
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

		var header = &fasthttp.RequestHeader{}
		md2Head(md, header)

		return nil, wrapperStream(ctx,
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
			&response{header: new(middleware2.ResponseHeader)},
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
