package gatewayserver

import (
	"context"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors/errcode"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/proto/errorpb"
	"github.com/pubgo/funk/v2/strutil"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/pkg/grpcutil"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"github.com/pubgo/lava/v2/pkg/lava"
	"github.com/pubgo/lava/v2/pkg/lavacontexts"
)

// handlerRPCMiddle adapts lava Middleware to gateway.RPCMiddleware so unary and
// all streaming modes share one chain for both in-process and proxy backends.
func handlerRPCMiddle(middlewares map[string][]lava.Middleware) gateway.RPCMiddleware {
	return func(ctx context.Context, op *gateway.Operation, next gateway.RPCHandler) (header, trailer metadata.MD, err error) {
		if op == nil {
			return next(ctx)
		}

		reqMetadata := getIncomingMetadata(ctx).Copy()

		ct := defaultContentType
		if c := reqMetadata.Get("x-content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}
		if c := reqMetadata.Get("content-type"); len(c) != 0 && c[0] != "" {
			ct = c[0]
		}
		delete(reqMetadata, "x-content-type")

		if p, ok := peer.FromContext(ctx); ok {
			reqMetadata.Set("remote", p.Addr.String())
		}

		to := reqMetadata.Get("timeout")
		delete(reqMetadata, "timeout")
		if len(to) != 0 && to[0] != "" {
			dur, parseErr := time.ParseDuration(to[0])
			if parseErr != nil {
				log.Warn().
					Err(parseErr).
					Str("operation", op.FullMethod).
					Str("timeout", to[0]).
					Msg("invalid timeout metadata, running without request timeout")
			} else {
				dur = grpcutil.CapRequestTimeout(dur)
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, dur)
				defer cancel()
			}
		}

		url := op.FullMethod
		if _url, ok := reqMetadata["url"]; ok {
			url = _url[0]
		}

		reqHeader := httputil.NewRequestHeader()
		for k, v := range reqMetadata {
			for i := range v {
				reqHeader.Add(k, v[i])
			}
		}
		rspHeader := httputil.NewResponseHeader()

		srvName := serviceFromMethod(op.FullMethod)
		rpcReq := &rpcRequest{
			service:     srvName,
			method:      op.FullMethod,
			url:         url,
			contentType: ct,
			payload:     gateway.IncomingPayload(ctx),
			header:      reqHeader,
			rspHeader:   rspHeader,
		}
		if op.StreamDesc != nil {
			rpcReq.stream = stubStream{ctx: ctx}
		}

		reqId := strutil.FirstFnNotEmpty(
			func() string { return lavacontexts.GetReqID(ctx) },
			func() string { return string(rpcReq.Header().Peek(httputil.HeaderXRequestID)) },
			func() string { return xid.New().String() },
		)
		ctx = lavacontexts.CreateCtxWithReqID(ctx, reqId)
		reqHeader.Set(httputil.HeaderXRequestID, reqId)
		reqHeader.Set(httputil.HeaderXRequestVersion, version.Version())

		ctx = lavacontexts.CreateReqHeader(ctx, reqHeader)
		ctx = lavacontexts.CreateRspHeader(ctx, rspHeader)

		wrapper := func(ctx context.Context, _ lava.Request) (lava.Response, error) {
			var nextErr error
			header, trailer, nextErr = next(ctx)
			if nextErr != nil {
				return nil, nextErr
			}
			return &rpcResponse{header: rspHeader}, nil
		}

		_, err = lava.Chain(middlewares[srvName]...).Middleware(wrapper)(ctx, rpcReq)
		if err != nil {
			pb := errcode.ParseError(err)
			if pb.Message == "" {
				pb.Message = err.Error()
			}
			if pb.Code == 0 {
				pb.StatusCode = errorpb.Code_Internal
				pb.Code = int32(errcode.GrpcCodeToHTTP(codes.Code(errorpb.Code_Internal)))
			}
			log.Error().
				Err(err).
				Str("operation", op.FullMethod).
				Str("request_id", reqId).
				Int32("code", pb.Code).
				Msg("rpc middleware rejected request")
			return nil, nil, errcode.ConvertErr2Status(pb).Err()
		}

		if header == nil {
			header = metadata.MD{}
		}
		header.Set(httputil.HeaderXRequestID, reqId)
		header.Set(httputil.HeaderXRequestVersion, version.Version())
		header.Set(httputil.HeaderXRequestOperation, op.FullMethod)
		rspHeader.VisitAll(func(key, value []byte) {
			header.Append(string(key), string(value))
		})
		return header, trailer, nil
	}
}

// stubStream lets rpcRequest.Stream() report streaming RPCs without a real
// ServerStream in the middleware layer (the pump runs inside next()).
type stubStream struct{ ctx context.Context }

func (s stubStream) SetHeader(metadata.MD) error  { return nil }
func (s stubStream) SendHeader(metadata.MD) error { return nil }
func (s stubStream) SetTrailer(metadata.MD)       {}
func (s stubStream) Context() context.Context     { return s.ctx }
func (s stubStream) SendMsg(any) error            { return nil }
func (s stubStream) RecvMsg(any) error            { return nil }

var _ grpc.ServerStream = stubStream{}
