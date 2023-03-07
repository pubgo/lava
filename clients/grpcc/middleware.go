package grpcc

import (
	"context"
	"strings"
	"time"

	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/strutil"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/pkg/grpcutil"
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
	var unaryWrapper = func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var md = make(metadata.MD)
		head2md(req.Header(), md)
		ctx = metadata.NewOutgoingContext(ctx, md)
		var reqCtx = req.(*request)
		var header = make(metadata.MD)
		var trailer = make(metadata.MD)
		reqCtx.opts = append(reqCtx.opts, grpc.Header(&header), grpc.Trailer(&trailer))

		if err := reqCtx.invoker(ctx, reqCtx.method, reqCtx.req, reqCtx.resp, reqCtx.cc, reqCtx.opts...); err != nil {
			return nil, err
		}

		rsp := &response{resp: reqCtx.resp, header: new(lava.ResponseHeader)}
		md2Head(header, rsp.header)
		md2Head(trailer, rsp.header)
		return rsp, nil
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

		var header = &fasthttp.RequestHeader{}
		md2Head(md, header)

		_, err = unaryWrapper(ctx, &request{
			ct:      ct,
			header:  header,
			service: serviceFromMethod(method),
			opts:    opts,
			method:  method,
			req:     req,
			cc:      cc,
			invoker: invoker,
		})
		return err
	}
}

func streamInterceptor(middlewares []lava.Middleware) grpc.StreamClientInterceptor {
	wrapperStream := func(ctx context.Context, req lava.Request) (lava.Response, error) {
		var reqCtx = req.(*request)
		var md = make(metadata.MD)
		head2md(req.Header(), md)

		ctx = metadata.NewOutgoingContext(ctx, md)
		stream, err := reqCtx.streamer(ctx, reqCtx.desc, reqCtx.cc, reqCtx.method, reqCtx.opts...)
		if err != nil {
			return nil, err
		}

		return &response{header: new(lava.ResponseHeader), stream: stream}, nil
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

		var header = &fasthttp.RequestHeader{}
		md2Head(md, header)
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
