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
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/pubgo/lava/v2/pkg/gateway/internal"
	"github.com/pubgo/lava/v2/pkg/gateway/routertree"
	"github.com/pubgo/lava/v2/pkg/lava"
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

// WithCodec registers or overrides a Codec for the given Content-Type
// (e.g. "application/json"). Built-in defaults cover JSON and Protobuf.
func WithCodec(contentType string, c Codec) MuxOption {
	return func(o *muxOptions) {
		if o.codecs == nil {
			o.codecs = make(map[string]Codec)
		}
		o.codecs[contentType] = c
	}
}

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
	regErr       error

	// backendUnaryInts / backendStreamInts wrap Invoke/NewStream for both
	// in-process and proxy backends (see UseBackend*).
	backendUnaryInts  []BackendUnaryInterceptor
	backendStreamInts []BackendStreamInterceptor

	// rpcMiddleware wraps full Dispatch / DispatchFrontend (see UseRPCMiddleware).
	rpcMiddleware []RPCMiddleware
}

// Err returns the first service registration error, if any.
func (m *Mux) Err() error { return m.regErr }

func (m *Mux) setRegErr(err error) {
	if err != nil && m.regErr == nil {
		m.regErr = err
	}
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
	act := m.findMethodByName(name)
	if act == nil {
		return nil
	}

	return handleOperation(act)
}

func (m *Mux) GetOperation(operation string) *GrpcMethod {
	opt := m.findMethod(operation)
	if opt == nil {
		return nil
	}

	return handleOperation(opt)
}

// LookupOperation returns the registered Operation for a gRPC full method, or nil.
func (m *Mux) LookupOperation(fullMethod string) *Operation {
	return operationFromMethod(m.findMethod(fullMethod))
}

// findMethod returns the registration record for a gRPC full method.
// Prefer LookupOperation for dispatch; this stays package-private for codec /
// proxy binding that still needs methodWrapper.
func (m *Mux) findMethod(fullMethod string) *methodWrapper {
	if m == nil || m.opts == nil {
		return nil
	}
	return m.opts.handlers[fullMethod]
}

func (m *Mux) findMethodByName(name string) *methodWrapper {
	if m == nil || m.opts == nil {
		return nil
	}
	return m.opts.customOperationNames[name]
}

func (m *Mux) Handler(ctx fiber.Ctx) error {
	return m.httpFrontend.handle(ctx)
}

func applyResponseMetadata(ctx fiber.Ctx, md metadata.MD) {
	applyResponseMetadataOpts(ctx, md, false)
}

// applyGRPCWebMetadata writes metadata into the Fiber response, including
// grpc-* keys so fiberWebWriter can emit them as a gRPC-Web trailer frame.
func applyGRPCWebMetadata(ctx fiber.Ctx, md metadata.MD) {
	applyResponseMetadataOpts(ctx, md, true)
}

func applyResponseMetadataOpts(ctx fiber.Ctx, md metadata.MD, allowGRPCKeys bool) {
	for k, v := range md {
		kLower := strings.ToLower(k)
		if isReservedHeader(kLower) && !isWhitelistedHeader(kLower) {
			if !(allowGRPCKeys && strings.HasPrefix(kLower, "grpc-")) {
				continue
			}
		}
		v = lo.Filter(v, func(item string, index int) bool { return item != "" })
		if len(v) == 0 {
			continue
		}
		if strings.HasSuffix(kLower, binHdrSuffix) {
			encoded := make([]string, len(v))
			for i, item := range v {
				encoded[i] = encodeBinHeader([]byte(item))
			}
			v = encoded
		}
		ctx.Response().Header.Set(k, v[0])
		for i := 1; i < len(v); i++ {
			ctx.Response().Header.Add(k, v[i])
		}
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
	invoker := chainBackendUnary(m.backendUnaryInts, m.invokeBackend)
	return invoker(ctx, method, args, reply, opts...)
}

func (m *Mux) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	streamer := chainBackendStream(m.backendStreamInts, m.newBackendStream)
	return streamer(ctx, desc, method, opts...)
}

func (m *Mux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// ServeHTTP exposes only the HTTP/REST + gRPC-Web frontend (Fiber Handler).
	// WebSocket requires Mux.WebSocketHandler on a dedicated net/http server;
	// native gRPC requires GRPCServerOptions / UnknownServiceHandler.
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

	// Compressors are registered and applied on the HTTP/gRPC-Web framed path
	// (grpc-encoding / grpc-accept-encoding + per-frame compression flag).
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
	// Compatibility: wraps RegisterService handlers inside inprocgrpc only.
	// Prefer UseBackendUnaryInterceptor for middleware that must also cover RegisterProxy.
	m.localClient.WithServerUnaryInterceptor(interceptor)
}

// SetStreamInterceptor configures the in-process channel to use the
// given server interceptor for streaming RPCs when dispatching.
// Prefer UseBackendStreamInterceptor for middleware that must also cover RegisterProxy.
func (m *Mux) SetStreamInterceptor(interceptor grpc.StreamServerInterceptor) {
	m.localClient.WithServerStreamInterceptor(interceptor)
}

func (m *Mux) RegisterProxy(sd *grpc.ServiceDesc, proxy lava.GrpcRouter, cli grpc.ClientConnInterface) {
	assert.If(cli == nil, "cli is nil")
	if err := m.registerService(sd, proxy, cli); err != nil {
		m.setRegErr(errors.Wrapf(err, "gateway: RegisterProxy %s", sd.ServiceName))
	}
}

// RegisterService satisfies grpc.ServiceRegistrar for generated service code hooks.
func (m *Mux) RegisterService(sd *grpc.ServiceDesc, ss any) {
	assert.If(funk.IsNil(ss), "ss params is nil")

	m.localClient.RegisterService(sd, ss)

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		m.setRegErr(errors.Errorf("gateway: RegisterService handler type %v does not satisfy %v", st, ht))
		return
	}

	if err := m.registerService(sd, ss, nil); err != nil {
		m.setRegErr(errors.Wrapf(err, "gateway: RegisterService %s", sd.ServiceName))
	}
}

func (m *Mux) registerRouter(rule *methodWrapper) {
	m.opts.handlers[rule.grpcFullMethod] = rule
	if rule.meta != nil {
		assert.If(m.opts.customOperationNames[rule.meta.Name] != nil, "rpc custom name:%s already exists", rule.meta.Name)
		m.opts.customOperationNames[rule.meta.Name] = rule
	}

	rule.inputType = assert.Must1(protoregistry.GlobalTypes.FindMessageByName(rule.grpcMethodProtoDesc.Input().FullName()))
	rule.outputType = assert.Must1(protoregistry.GlobalTypes.FindMessageByName(rule.grpcMethodProtoDesc.Output().FullName()))
	rule.op = &Operation{
		FullMethod: rule.grpcFullMethod,
		InputType:  rule.inputType,
		OutputType: rule.outputType,
		StreamDesc: rule.grpcStreamDesc,
		Meta:       rule.meta,
	}

	// HTTP index only: path → FullMethod. Schema lives on Operation.
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

// GetRouterTarget matches an HTTP method and path against the mux route tree.
// method defaults to POST when empty (the method used for auto-registered gRPC
// full-method routes).
func GetRouterTarget(mux *Mux, method, path string) (*MatchOperation, error) {
	if path == "" {
		return nil, errors.New("path is null")
	}

	if method == "" {
		method = http.MethodPost
	}

	restTarget, err := mux.routerTree.Match(method, path)
	if err != nil {
		return nil, errors.Wrapf(err, "path not found, method=%s path=%s", method, path)
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

// Routes returns registered Operations for external bridges (e.g. zrpc).
func (m *Mux) Routes() []MethodRoute {
	routes := make([]MethodRoute, 0, len(m.opts.handlers))
	for fullMethod, mth := range m.opts.handlers {
		if op := operationFromMethod(mth); op != nil {
			routes = append(routes, MethodRoute{FullMethod: fullMethod, Operation: op})
		}
	}
	return routes
}

// Dispatch routes a frontend stream to the Mux backend (through RPC middleware).
func (m *Mux) Dispatch(ctx context.Context, frontend FrontendStream, op *Operation, in any, opts ...DispatchOption) (metadata.MD, metadata.MD, error) {
	if in != nil {
		ctx = context.WithValue(ctx, rpcIncomingPayloadKey{}, in)
	}
	return m.runRPC(ctx, op, func(ctx context.Context) (metadata.MD, metadata.MD, error) {
		return m.dispatcher.Dispatch(ctx, m, frontend, op, in, opts...)
	})
}

// DispatchFrontend drives a frontend stream end-to-end. It pre-reads the request
// for unary/server-stream, then runs RPC middleware around Dispatch so middleware
// sees the payload and the full stream lifetime for all modes.
func (m *Mux) DispatchFrontend(ctx context.Context, frontend FrontendStream, op *Operation, opts ...DispatchOption) (metadata.MD, metadata.MD, error) {
	if op == nil {
		return m.dispatcher.DispatchFrontend(ctx, m, frontend, op, opts...)
	}

	preReadRequest := op.StreamDesc == nil ||
		(op.StreamDesc.ServerStreams && !op.StreamDesc.ClientStreams)

	var in any
	if preReadRequest {
		req := op.InputType.New().Interface()
		if err := frontend.RecvMsg(req); err != nil {
			return nil, nil, err
		}
		in = req
	}

	return m.Dispatch(ctx, frontend, op, in, opts...)
}
