package gateway

import (
	"context"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
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

func grpcMethodHandlerWrapper(mth *methodWrapper, opts ...grpc.CallOption) MethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		in := mth.inputType.New().Interface()
		if err := dec(in); err != nil {
			return nil, errors.WrapCaller(err)
		}

		h := func(ctx context.Context, req any) (any, error) {
			out := mth.outputType.New().Interface()
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
