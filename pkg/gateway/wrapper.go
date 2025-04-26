package gateway

import (
	"context"

	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/pkg/proto/lavapbv1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type serviceWrapper struct {
	opts           *muxOptions
	srv            any
	serviceDesc    *grpc.ServiceDesc
	servicePbDesc  protoreflect.ServiceDescriptor
	remoteProxyCli grpc.ClientConnInterface
}

type GrpcMethod struct {
	Srv     any
	SrvDesc *grpc.ServiceDesc

	GrpcMethodDesc *grpc.MethodDesc
	GrpcStreamDesc *grpc.StreamDesc
	MethodDesc     protoreflect.MethodDescriptor

	GrpcFullMethod string
	Meta           *lavapbv1.RpcMeta
}

type methodWrapper struct {
	srv                 *serviceWrapper
	grpcMethodDesc      *grpc.MethodDesc
	grpcStreamDesc      *grpc.StreamDesc
	grpcMethodProtoDesc protoreflect.MethodDescriptor

	inputType  protoreflect.MessageType
	outputType protoreflect.MessageType

	// /{ServiceName}/{MethodName}
	grpcFullMethod string
	meta           *lavapbv1.RpcMeta
}

//func (h methodWrapper) Handle(stream grpc.ServerStream) error {
//	if h.grpcMethodDesc != nil {
//		ctx := stream.Context()
//
//		reply, err := h.grpcMethodDesc.Exec(h.srv.srv, ctx, stream.RecvMsg, h.srv.opts.unaryInterceptor)
//		if err != nil {
//			return errors.WrapCaller(err)
//		}
//
//		return errors.WrapCaller(stream.SendMsg(reply))
//	} else if h.grpcStreamDesc != nil {
//		info := &grpc.StreamServerInfo{
//			FullMethod:     h.grpcFullMethod,
//			IsClientStream: h.grpcStreamDesc.ClientStreams,
//			IsServerStream: h.grpcStreamDesc.ServerStreams,
//		}
//
//		if h.srv.opts.streamInterceptor != nil {
//			return errors.WrapCaller(h.srv.opts.streamInterceptor(h.srv.srv, stream, info, h.grpcStreamDesc.Exec))
//		} else {
//			return errors.WrapCaller(h.grpcStreamDesc.Exec(h.srv.srv, stream))
//		}
//	} else {
//		return errors.Format("cannot find server handler")
//	}
//}

func grpcMethodHandlerWrapper(mth *methodWrapper, opts ...grpc.CallOption) GrpcMethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		var in = mth.inputType.New().Interface()
		if err := dec(in); err != nil {
			return nil, errors.WrapCaller(err)
		}

		var h = func(ctx context.Context, req any) (any, error) {
			var out = mth.outputType.New().Interface()
			err := mth.srv.remoteProxyCli.Invoke(ctx, mth.grpcFullMethod, req, out, opts...)
			if err != nil {
				return nil, err
			}
			return out, nil
		}

		// 获取 server header 并转换成 client header
		if interceptor == nil {
			return h(ctx, in)
		}

		return interceptor(ctx, in, &grpc.UnaryServerInfo{FullMethod: mth.grpcFullMethod}, h)
	}
}
