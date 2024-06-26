package gateway

import (
	"context"
	"github.com/pubgo/funk/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type serviceWrapper struct {
	opts          *muxOptions
	srv           interface{}
	target        grpc.ClientConnInterface
	serviceDesc   *grpc.ServiceDesc
	servicePbDesc protoreflect.ServiceDescriptor
}

type methodWrapper struct {
	srv              *serviceWrapper
	grpcMethodDesc   *grpc.MethodDesc
	grpcStreamDesc   *grpc.StreamDesc
	grpcMethodPbDesc protoreflect.MethodDescriptor

	// /{ServiceName}/{MethodName}
	grpcFullMethod string
}

func (h methodWrapper) Handle(stream grpc.ServerStream) error {
	if h.grpcMethodDesc != nil {
		ctx := stream.Context()

		if h.srv.target != nil {
			h.srv.opts.unaryInterceptor(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
				// change header
				h.srv.target.Invoke(stream.Context(), "", nil, nil)
			})
		} else {
			reply, err := h.grpcMethodDesc.Handler(h.srv.srv, ctx, stream.RecvMsg, h.srv.opts.unaryInterceptor)
			if err != nil {
				return errors.WrapCaller(err)
			}

			return errors.WrapCaller(stream.SendMsg(reply))
		}

	} else {
		info := &grpc.StreamServerInfo{
			FullMethod:     h.grpcFullMethod,
			IsClientStream: h.grpcStreamDesc.ClientStreams,
			IsServerStream: h.grpcStreamDesc.ServerStreams,
		}

		if h.srv.target != nil {
			clientStream, err := h.srv.target.NewStream(nil, h.grpcStreamDesc, "", nil)

			h.srv.opts.streamInterceptor(nil, stream, info, func(srv any, stream grpc.ServerStream) error {
				_ = stream
				_ = clientStream

				go func() {
					for {
						var msg any
						clientStream.RecvMsg(&msg)
						stream.SendMsg(msg)
					}
				}()

				for {
					var msg any
					stream.RecvMsg(&msg)
					clientStream.SendMsg(msg)
				}
			})
		}

		if h.srv.opts.streamInterceptor != nil {
			return errors.WrapCaller(h.srv.opts.streamInterceptor(h.srv.srv, stream, info, h.grpcStreamDesc.Handler))
		} else {
			return errors.WrapCaller(h.grpcStreamDesc.Handler(h.srv.srv, stream))
		}
	}
}
