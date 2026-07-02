package gateway

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/pubgo/lava/v2/lava"
	"github.com/pubgo/lava/v2/pkg/gateway/internal"
	"github.com/pubgo/lava/v2/pkg/gateway/routertree"
)

type muxOptions struct {
	types                protoregistry.MessageTypeResolver
	files                *protoregistry.Files
	codecs               map[string]Codec
	codecsByName         map[string]Codec
	compressors          map[string]Compressor
	requestInterceptors  map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error
	responseInterceptors map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error
	handlers             map[string]*methodWrapper
	customOperationNames map[string]*methodWrapper
}

// MuxOption is an option for a mux.
type MuxOption func(*muxOptions)

var (
	defaultMuxOptions = muxOptions{
		files:                protoregistry.GlobalFiles,
		types:                protoregistry.GlobalTypes,
		responseInterceptors: make(map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error),
		requestInterceptors:  make(map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error),
		handlers:             make(map[string]*methodWrapper),
		customOperationNames: make(map[string]*methodWrapper),
	}

	defaultCodecs = map[string]Codec{
		"application/json":         CodecJSON{},
		"application/protobuf":     CodecProto{},
		"application/octet-stream": CodecProto{},
		"google.api.HttpBody":      codecHTTPBody{},
	}

	defaultCompressors = map[string]Compressor{
		"gzip":     &internal.CompressorGzip{},
		"identity": nil,
	}
)

var _ Gateway = (*Mux)(nil)

type Mux struct {
	localClient  *inprocgrpc.Channel
	opts         *muxOptions
	routerTree   *routertree.RouteTree
	dispatcher   *Dispatcher
	httpFrontend *httpFrontend
}

func (m *Mux) GetRouteMethods() []RouteOperation { return m.routerTree.List() }

func (m *Mux) SetResponseEncoder(name protoreflect.FullName, f func(ctx fiber.Ctx, msg proto.Message) error) {
	m.opts.responseInterceptors[name] = f
}

func (m *Mux) SetRequestDecoder(name protoreflect.FullName, f func(ctx fiber.Ctx, msg proto.Message) error) {
	m.opts.requestInterceptors[name] = f
}

func (m *Mux) MatchOperation(method, path string) (r result.Result[*MatchOperation]) {
	return result.Wrap(m.routerTree.Match(method, path)).
		Log(func(e result.Event) {
			e.Str("method", method)
			e.Str("path", path)
			e.Msg("match operation failed")
		})
}

func (m *Mux) GetOperationByName(name string) *GrpcMethod {
	act := m.opts.customOperationNames[name]
	if act == nil {
		return nil
	}

	return handleOperation(act)
}

func (m *Mux) GetOperation(operation string) *GrpcMethod {
	opt := m.opts.handlers[operation]
	if opt == nil {
		return nil
	}

	return handleOperation(opt)
}

func (m *Mux) Handler(ctx fiber.Ctx) error {
	return m.httpFrontend.handle(ctx)
}

func (m *Mux) invokeWithStream(stream *streamHTTP, in any) error {
	header, trailer, err := m.dispatcher.Dispatch(stream.Context(), m, stream, operationFromMethod(stream.method), in)
	if err != nil {
		return err
	}
	applyResponseMetadata(stream.handler, header)
	applyResponseMetadata(stream.handler, trailer)
	applyResponseMetadata(stream.handler, stream.trailer)
	return nil
}

func (m *Mux) invokeResponseStream(remoteStream *streamHTTP, in any) error {
	return m.invokeWithStream(remoteStream, in)
}

func applyResponseMetadata(ctx fiber.Ctx, md metadata.MD) {
	for k, v := range md {
		v = lo.Filter(v, func(item string, index int) bool { return item != "" })
		if len(v) == 0 {
			continue
		}
		ctx.Response().Header.Set(k, v[0])
	}
}

func isDuplicateHeaderError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "headers already sent") ||
		strings.Contains(msg, "sendheader called multiple times") ||
		strings.Contains(msg, "header already sent")
}

func (m *Mux) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	if mth := m.opts.handlers[method]; mth != nil {
		if mth.srv.remoteProxyCli != nil {
			return mth.srv.remoteProxyCli.Invoke(ctx, method, args, reply, opts...)
		}
	}

	return m.localClient.Invoke(ctx, method, args, reply, opts...)
}

func (m *Mux) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if mth := m.opts.handlers[method]; mth != nil {
		if mth.srv.remoteProxyCli != nil {
			return mth.srv.remoteProxyCli.NewStream(ctx, desc, method, opts...)
		}
	}

	return m.localClient.NewStream(ctx, desc, method, opts...)
}

func (m *Mux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// ServeHTTP acts as a thin wrapper only.
	// All protocol/business handling is centralized in Handler.
	adaptor.FiberHandler(m.Handler).ServeHTTP(writer, request)
}

func NewMux(opts ...MuxOption) *Mux {
	muxOpts := muxOptions{
		files:                defaultMuxOptions.files,
		types:                defaultMuxOptions.types,
		codecs:               defaultMuxOptions.codecs,
		codecsByName:         defaultMuxOptions.codecsByName,
		compressors:          defaultMuxOptions.compressors,
		handlers:             make(map[string]*methodWrapper),
		customOperationNames: make(map[string]*methodWrapper),
		requestInterceptors:  make(map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error),
		responseInterceptors: make(map[protoreflect.FullName]func(ctx fiber.Ctx, msg proto.Message) error),
	}
	for k, v := range defaultMuxOptions.requestInterceptors {
		muxOpts.requestInterceptors[k] = v
	}
	for k, v := range defaultMuxOptions.responseInterceptors {
		muxOpts.responseInterceptors[k] = v
	}
	for _, opt := range opts {
		opt(&muxOpts)
	}

	// Ensure codecs are set.
	if muxOpts.codecs == nil {
		muxOpts.codecs = make(map[string]Codec)
	}

	for k, v := range defaultCodecs {
		if _, ok := muxOpts.codecs[k]; !ok {
			muxOpts.codecs[k] = v
		}
	}

	muxOpts.codecsByName = make(map[string]Codec)
	for _, v := range muxOpts.codecs {
		muxOpts.codecsByName[v.Name()] = v
	}

	// Ensure compressors are set.
	if muxOpts.compressors == nil {
		muxOpts.compressors = make(map[string]Compressor)
	}

	for k, v := range defaultCompressors {
		if _, ok := muxOpts.compressors[k]; !ok {
			muxOpts.compressors[k] = v
		}
	}

	mux := &Mux{
		opts:        &muxOpts,
		localClient: new(inprocgrpc.Channel),
		routerTree:  routertree.New(),
		dispatcher:  NewDispatcher(),
	}
	mux.httpFrontend = newHTTPFrontend(mux)

	return mux
}

func (m *Mux) SetUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) {
	m.localClient.WithServerUnaryInterceptor(interceptor)
}

// SetStreamInterceptor configures the in-process channel to use the
// given server interceptor for streaming RPCs when dispatching.
func (m *Mux) SetStreamInterceptor(interceptor grpc.StreamServerInterceptor) {
	m.localClient.WithServerStreamInterceptor(interceptor)
}

func (m *Mux) RegisterProxy(sd *grpc.ServiceDesc, proxy lava.GrpcRouter, cli grpc.ClientConnInterface) {
	assert.If(cli == nil, "cli is nil")
	if err := m.registerService(sd, proxy, cli); err != nil {
		log.Fatal().Err(err).Msgf("gateway: RegisterProxy error: %v", err)
	}
}

// RegisterService satisfies grpc.ServiceRegistrar for generated service code hooks.
func (m *Mux) RegisterService(sd *grpc.ServiceDesc, ss any) {
	assert.If(funk.IsNil(ss), "ss params is nil")

	m.localClient.RegisterService(sd, ss)

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		log.Fatal().Msgf("gateway: RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	if err := m.registerService(sd, ss, nil); err != nil {
		log.Fatal().Err(err).Msgf("gateway: RegisterService error: %v", err)
	}
}

func (m *Mux) registerRouter(rule *methodWrapper) {
	m.opts.handlers[rule.grpcFullMethod] = rule
	if rule.meta != nil {
		assert.If(m.opts.customOperationNames[rule.meta.Name] != nil, "rpc custome name:%s already exists", rule.meta.Name)
		m.opts.customOperationNames[rule.meta.Name] = rule
	}

	rule.inputType = assert.Must1(protoregistry.GlobalTypes.FindMessageByName(rule.grpcMethodProtoDesc.Input().FullName()))
	rule.outputType = assert.Must1(protoregistry.GlobalTypes.FindMessageByName(rule.grpcMethodProtoDesc.Output().FullName()))

	assert.Exit(m.routerTree.Add(
		http.MethodPost,
		rule.grpcFullMethod,
		rule.grpcFullMethod,
		resolveBodyDesc(rule.grpcMethodProtoDesc, "*", "*")),
	)
}

func (m *Mux) registerService(gsd *grpc.ServiceDesc, ss any, cli grpc.ClientConnInterface) error {
	d, err := m.opts.files.FindDescriptorByName(protoreflect.FullName(gsd.ServiceName))
	if err != nil {
		return errors.WrapCaller(err)
	}

	sd, ok := d.(protoreflect.ServiceDescriptor)
	if !ok {
		return errors.Errorf("invalid httpPathRule descriptor %T", d)
	}

	srv := &serviceWrapper{
		opts:           m.opts,
		srv:            ss,
		serviceDesc:    gsd,
		servicePbDesc:  sd,
		remoteProxyCli: cli,
	}

	findMethodDesc := func(methodName string) protoreflect.MethodDescriptor {
		md := sd.Methods().ByName(protoreflect.Name(methodName))
		assert.If(md == nil, "missing protobuf descriptor for %v", methodName)
		return md
	}

	for i := range gsd.Methods {
		grpcMth := &gsd.Methods[i]
		methodDesc := findMethodDesc(grpcMth.MethodName)

		grpcMethod := fmt.Sprintf("/%s/%s", gsd.ServiceName, grpcMth.MethodName)
		assert.If(m.opts.handlers[grpcMethod] != nil, "grpc httpPathRule has existed")

		m.registerRouter(&methodWrapper{
			srv:                 srv,
			grpcMethodDesc:      grpcMth,
			grpcMethodProtoDesc: methodDesc,
			grpcFullMethod:      grpcMethod,
			meta:                getExtensionRpc(methodDesc),
		})

		assert.Exit(handlerHttpRoute(getExtensionHTTP(methodDesc), func(mth, path, reqBody, rspBody string) error {
			return errors.WrapCaller(m.routerTree.Add(mth, path, grpcMethod, resolveBodyDesc(methodDesc, reqBody, rspBody)))
		}))
	}

	for i := range gsd.Streams {
		grpcMth := &gsd.Streams[i]
		grpcMethod := "/" + gsd.ServiceName + "/" + grpcMth.StreamName
		assert.If(m.opts.handlers[grpcMethod] != nil, "grpc httpPathRule has existed")

		methodDesc := findMethodDesc(grpcMth.StreamName)

		m.registerRouter(&methodWrapper{
			srv:                 srv,
			grpcStreamDesc:      grpcMth,
			grpcMethodProtoDesc: methodDesc,
			grpcFullMethod:      grpcMethod,
			meta:                getExtensionRpc(methodDesc),
		})

		assert.Exit(handlerHttpRoute(getExtensionHTTP(methodDesc), func(mth, path, reqBody, rspBody string) error {
			return errors.WrapCaller(m.routerTree.Add(mth, path, grpcMethod, resolveBodyDesc(methodDesc, reqBody, rspBody)))
		}))
	}

	return nil
}

func GetRouterTarget(mux *Mux, kind, path string) (*MatchOperation, error) {
	if path == "" {
		return nil, errors.New("path is null")
	}

	if kind == "" {
		kind = "ws"
	}

	restTarget, err := mux.routerTree.Match(path, kind)
	if err != nil {
		return nil, errors.Wrapf(err, "path not found, kind=%s path=%s", kind, path)
	}

	return restTarget, nil
}

func handleOperation(opt *methodWrapper) *GrpcMethod {
	return &GrpcMethod{
		Srv:            opt.srv.srv,
		SrvDesc:        opt.srv.serviceDesc,
		GrpcMethodDesc: opt.grpcMethodDesc,
		GrpcStreamDesc: opt.grpcStreamDesc,
		MethodDesc:     opt.grpcMethodProtoDesc,
		GrpcFullMethod: opt.grpcFullMethod,
		Meta:           opt.meta,
	}
}

// Routes returns registered gRPC full methods for external bridges (e.g. zrpc).
func (m *Mux) Routes() []MethodRoute {
	routes := make([]MethodRoute, 0, len(m.opts.handlers))
	for fullMethod, mth := range m.opts.handlers {
		if op := operationFromMethod(mth); op != nil {
			routes = append(routes, MethodRoute{FullMethod: fullMethod, Operation: op})
		}
	}
	return routes
}

// Dispatch routes a frontend stream to the Mux backend.
func (m *Mux) Dispatch(ctx context.Context, frontend FrontendStream, op *Operation, in any) (metadata.MD, metadata.MD, error) {
	return m.dispatcher.Dispatch(ctx, m, frontend, op, in)
}

// DispatchFrontend drives a frontend stream end-to-end through the dispatcher.
func (m *Mux) DispatchFrontend(ctx context.Context, frontend FrontendStream, op *Operation) (metadata.MD, metadata.MD, error) {
	return m.dispatcher.DispatchFrontend(ctx, m, frontend, op)
}
