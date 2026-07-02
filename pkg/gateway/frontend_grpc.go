package gateway

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCPassthroughStreamHandler returns a stream handler suitable for
// grpc.UnknownServiceHandler. Native gRPC clients connect to a *grpc.Server
// configured with this handler; each RPC is forwarded to the Mux backend
// (inprocgrpc local handlers or remote proxies) through the unified Dispatcher,
// so unary, server-stream, client-stream, and bidi all behave correctly.
//
// Register services only on the Mux (RegisterService / RegisterProxy); do not
// register the same implementations again on the outer grpc.Server.
func (m *Mux) GRPCPassthroughStreamHandler() grpc.StreamHandler {
	return func(_ any, stream grpc.ServerStream) error {
		fullMethod, ok := grpc.MethodFromServerStream(stream)
		if !ok {
			return status.Error(codes.Internal, "gateway: method not found in stream context")
		}

		mth := m.opts.handlers[fullMethod]
		if mth == nil {
			return status.Errorf(codes.Unimplemented, "unknown method: %s", fullMethod)
		}

		op := operationFromMethod(mth)
		header, trailer, err := m.dispatcher.DispatchFrontend(stream.Context(), m, stream, op)
		if len(header) > 0 {
			_ = stream.SetHeader(header)
		}
		if len(trailer) > 0 {
			stream.SetTrailer(trailer)
		}
		return err
	}
}

// GRPCServerOptions returns server options that enable native gRPC passthrough
// through this Mux. Append to grpc.NewServer or grpcbuilder.Config.Build opts.
func (m *Mux) GRPCServerOptions(extra ...grpc.ServerOption) []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0, 1+len(extra))
	opts = append(opts, grpc.UnknownServiceHandler(m.GRPCPassthroughStreamHandler()))
	opts = append(opts, extra...)
	return opts
}
