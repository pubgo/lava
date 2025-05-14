package grpcs

import (
	"context"
	"fmt"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpccconfig"
	"github.com/pubgo/lava/core/metrics"
	"github.com/pubgo/lava/internal/middlewares/middleware_accesslog"
	"github.com/pubgo/lava/internal/middlewares/middleware_metric"
	"github.com/pubgo/lava/internal/middlewares/middleware_recovery"
	"github.com/pubgo/lava/internal/middlewares/middleware_service_info"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// NewInner grpc 服务内部通信
func NewInner(handlers []lava.GrpcRouter, grpcProxy []lava.GrpcProxy, dixMiddlewares []lava.Middleware, metric metrics.Metric, log log.Logger) *lava.InnerServer {
	middlewares := lava.Middlewares{
		middleware_service_info.New(),
		middleware_metric.New(metric),
		middleware_accesslog.New(log),
		middleware_recovery.New(),
	}
	middlewares = append(middlewares, dixMiddlewares...)

	cc := new(inprocgrpc.Channel)
	srvMidMap := make(map[string][]lava.Middleware)
	for _, h := range handlers {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)
		cc.RegisterService(h.ServiceDesc(), h)
	}

	for _, h := range grpcProxy {
		desc := h.ServiceDesc()
		assert.If(desc == nil, "desc is nil")

		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], middlewares...)
		srvMidMap[desc.ServiceName] = append(srvMidMap[desc.ServiceName], h.Middlewares()...)

		cli := grpcc.New(
			&grpccconfig.Cfg{
				Service: &grpccconfig.ServiceCfg{
					Name:   h.Proxy().Name,
					Addr:   h.Proxy().Addr,
					Scheme: h.Proxy().Resolver,
				},
			},
			grpcc.Params{
				Log:    log,
				Metric: metric,
			},
			h.Middlewares()...,
		)

		for i := range desc.Methods {
			var fullPath = fmt.Sprintf("/%s/%s", desc.ServiceName, desc.Methods[i].MethodName)
			inT, outT := getMthType(desc.ServiceName, desc.Methods[i].MethodName)
			desc.Methods[i].Handler = grpcMethodHandlerWrapper(cli, fullPath, inT, outT)
		}

		for i := range desc.Streams {
			inT, outT := getMthType(desc.ServiceName, desc.Methods[i].MethodName)
			desc.Streams[i].Handler = grpcMethodStreamWrapper(cli, inT, outT)
		}
		cc.RegisterService(h.ServiceDesc(), h)
	}

	cc = cc.WithServerUnaryInterceptor(handlerUnaryMiddle(srvMidMap))
	cc = cc.WithServerStreamInterceptor(handlerStreamMiddle(srvMidMap))
	return &lava.InnerServer{ClientConnInterface: cc}
}

func grpcMethodHandlerWrapper(cli grpc.ClientConnInterface, fullPath string, inType, outType protoreflect.MessageType) gateway.GrpcMethodHandler {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		var in = inType.New().Interface()
		if err := dec(in); err != nil {
			return nil, errors.WrapCaller(err)
		}

		var h = func(ctx context.Context, req any) (any, error) {
			var out = outType.New().Interface()
			var header metadata.MD
			var trailer metadata.MD
			err := cli.Invoke(ctx, fullPath, req, out, append([]grpc.CallOption{}, grpc.Header(&header), grpc.Trailer(&trailer))...)
			if err != nil {
				return nil, errors.WrapCaller(err)
			}
			return out, nil
		}

		// 获取 server header 并转换成 client header
		if interceptor == nil {
			return h(ctx, in)
		}

		return interceptor(ctx, in, &grpc.UnaryServerInfo{FullMethod: fullPath}, h)
	}
}

func getMthType(srvName string, mthName string) (protoreflect.MessageType, protoreflect.MessageType) {
	d := assert.Must1(protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(srvName)))

	sd, ok := d.(protoreflect.ServiceDescriptor)
	if !ok {
		assert.Must(errors.Format("invalid httpPathRule descriptor %T", d))
	}

	findMethodDesc := func(methodName string) protoreflect.MethodDescriptor {
		md := sd.Methods().ByName(protoreflect.Name(methodName))
		assert.If(md == nil, "missing protobuf descriptor for %v", methodName)
		return md
	}

	mthDesc := findMethodDesc(mthName)

	inputType := assert.Must1(protoregistry.GlobalTypes.FindMessageByName(mthDesc.Input().FullName()))
	outputType := assert.Must1(protoregistry.GlobalTypes.FindMessageByName(mthDesc.Output().FullName()))
	return inputType, outputType
}

func grpcMethodStreamWrapper(cli grpc.ClientConnInterface, inType, outType protoreflect.MessageType) gateway.GrpcStreamHandler {
	return gateway.TransparentHandler(cli, inType, outType)
}
