package gateway

import (
	"github.com/pubgo/funk/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type serviceWrap struct {
	opts        *muxOptions
	srv         interface{}
	serviceDesc *grpc.ServiceDesc
	servicePB   protoreflect.ServiceDescriptor
}

type methodWrap struct {
	srv        *serviceWrap
	methodDesc *grpc.MethodDesc
	streamDesc *grpc.StreamDesc
	grpcMethod protoreflect.MethodDescriptor

	// /{ServiceName}/{MethodName}
	grpcMethodName string
}

func (h methodWrap) Handle(stream grpc.ServerStream) error {
	if h.methodDesc != nil {
		ctx := stream.Context()

		reply, err := h.methodDesc.Handler(h.srv.srv, ctx, stream.RecvMsg, h.srv.opts.unaryInterceptor)
		if err != nil {
			return errors.WrapCaller(err)
		}

		return errors.WrapCaller(stream.SendMsg(reply))
	} else {
		info := &grpc.StreamServerInfo{
			FullMethod:     string(h.grpcMethod.FullName()),
			IsClientStream: h.streamDesc.ClientStreams,
			IsServerStream: h.streamDesc.ServerStreams,
		}

		if h.srv.opts.streamInterceptor != nil {
			return h.srv.opts.streamInterceptor(h.srv.srv, stream, info, h.streamDesc.Handler)
		} else {
			return h.streamDesc.Handler(h.srv.srv, stream)
		}
	}
}
