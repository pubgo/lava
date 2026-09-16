package gateway

import (
	"context"

	"google.golang.org/grpc"
)

// BackendUnaryInterceptor wraps Mux.Invoke for both in-process and proxy backends.
// Unlike grpc.UnaryClientInterceptor it is not tied to *grpc.ClientConn.
type BackendUnaryInterceptor func(
	ctx context.Context,
	method string,
	req, reply any,
	invoker func(context.Context, string, any, any, ...grpc.CallOption) error,
	opts ...grpc.CallOption,
) error

// BackendStreamInterceptor wraps Mux.NewStream for both in-process and proxy backends.
type BackendStreamInterceptor func(
	ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	streamer func(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error),
	opts ...grpc.CallOption,
) (grpc.ClientStream, error)

// UseBackendUnaryInterceptor appends an interceptor that wraps every Mux.Invoke,
// including both in-process and RegisterProxy backends.
func (m *Mux) UseBackendUnaryInterceptor(interceptor BackendUnaryInterceptor) {
	if interceptor == nil {
		return
	}
	m.backendUnaryInts = append(m.backendUnaryInts, interceptor)
}

// UseBackendStreamInterceptor appends an interceptor that wraps every Mux.NewStream,
// including both in-process and RegisterProxy backends.
func (m *Mux) UseBackendStreamInterceptor(interceptor BackendStreamInterceptor) {
	if interceptor == nil {
		return
	}
	m.backendStreamInts = append(m.backendStreamInts, interceptor)
}

func (m *Mux) invokeBackend(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	if mth := m.opts.handlers[method]; mth != nil && mth.srv.remoteProxyCli != nil {
		return mth.srv.remoteProxyCli.Invoke(ctx, method, args, reply, opts...)
	}
	return m.localClient.Invoke(ctx, method, args, reply, opts...)
}

func (m *Mux) newBackendStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if mth := m.opts.handlers[method]; mth != nil && mth.srv.remoteProxyCli != nil {
		return mth.srv.remoteProxyCli.NewStream(ctx, desc, method, opts...)
	}
	return m.localClient.NewStream(ctx, desc, method, opts...)
}

func chainBackendUnary(ints []BackendUnaryInterceptor, final func(context.Context, string, any, any, ...grpc.CallOption) error) func(context.Context, string, any, any, ...grpc.CallOption) error {
	invoker := final
	for i := len(ints) - 1; i >= 0; i-- {
		interceptor := ints[i]
		next := invoker
		invoker = func(ctx context.Context, method string, req, reply any, opts ...grpc.CallOption) error {
			return interceptor(ctx, method, req, reply, next, opts...)
		}
	}
	return invoker
}

func chainBackendStream(
	ints []BackendStreamInterceptor,
	final func(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error),
) func(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	streamer := final
	for i := len(ints) - 1; i >= 0; i-- {
		interceptor := ints[i]
		next := streamer
		streamer = func(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return interceptor(ctx, desc, method, next, opts...)
		}
	}
	return streamer
}
