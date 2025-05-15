package gateway

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/lava"
	"github.com/pubgo/lava/pkg/gateway/routertree"
	"github.com/pubgo/lava/pkg/httputil"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type muxOptions struct {
	types                 protoregistry.MessageTypeResolver
	files                 *protoregistry.Files
	codecs                map[string]Codec
	codecsByName          map[string]Codec
	compressors           map[string]Compressor
	contentTypeOffers     []string
	encodingTypeOffers    []string
	maxReceiveMessageSize int
	maxSendMessageSize    int
	connectionTimeout     time.Duration
	errHandler            func(err error, ctx *fiber.Ctx)
	requestInterceptors   map[protoreflect.FullName]func(ctx *fiber.Ctx, msg proto.Message) error
	responseInterceptors  map[protoreflect.FullName]func(ctx *fiber.Ctx, msg proto.Message) error
	handlers              map[string]*methodWrapper
	customOperationNames  map[string]*methodWrapper
}

// MuxOption is an option for a mux.
type MuxOption func(*muxOptions)

const (
	defaultServerMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultServerMaxSendMessageSize    = math.MaxInt32
	defaultServerConnectionTimeout     = 120 * time.Second
)

var (
	defaultMuxOptions = muxOptions{
		maxReceiveMessageSize: defaultServerMaxReceiveMessageSize,
		maxSendMessageSize:    defaultServerMaxSendMessageSize,
		connectionTimeout:     defaultServerConnectionTimeout,
		files:                 protoregistry.GlobalFiles,
		types:                 protoregistry.GlobalTypes,
		responseInterceptors:  make(map[protoreflect.FullName]func(ctx *fiber.Ctx, msg proto.Message) error),
		requestInterceptors:   make(map[protoreflect.FullName]func(ctx *fiber.Ctx, msg proto.Message) error),
		handlers:              make(map[string]*methodWrapper),
		customOperationNames:  make(map[string]*methodWrapper),
	}

	defaultCodecs = map[string]Codec{
		"application/json":         CodecJSON{},
		"application/protobuf":     CodecProto{},
		"application/octet-stream": CodecProto{},
		"google.api.HttpBody":      codecHTTPBody{},
	}

	defaultCompressors = map[string]Compressor{
		"gzip":     &CompressorGzip{},
		"identity": nil,
	}
)

func MaxReceiveMessageSizeOption(s int) MuxOption {
	return func(opts *muxOptions) { opts.maxReceiveMessageSize = s }
}

func MaxSendMessageSizeOption(s int) MuxOption {
	return func(opts *muxOptions) { opts.maxSendMessageSize = s }
}

func ConnectionTimeoutOption(d time.Duration) MuxOption {
	return func(opts *muxOptions) { opts.connectionTimeout = d }
}

func TypesOption(t protoregistry.MessageTypeResolver) MuxOption {
	return func(opts *muxOptions) { opts.types = t }
}

func FilesOption(f *protoregistry.Files) MuxOption {
	return func(opts *muxOptions) { opts.files = f }
}

// CodecOption registers a codec for the given content type.
func CodecOption(contentType string, c Codec) MuxOption {
	return func(opts *muxOptions) {
		if opts.codecs == nil {
			opts.codecs = make(map[string]Codec)
		}
		opts.codecs[contentType] = c
	}
}

// CompressorOption registers a compressor for the given content encoding.
func CompressorOption(contentEncoding string, c Compressor) MuxOption {
	return func(opts *muxOptions) {
		if opts.compressors == nil {
			opts.compressors = make(map[string]Compressor)
		}
		opts.compressors[contentEncoding] = c
	}
}

var _ Gateway = (*Mux)(nil)

type Mux struct {
	localClient *inprocgrpc.Channel
	opts        *muxOptions
	routerTree  *routertree.RouteTree
}

func (m *Mux) GetRouteMethods() []RouteOperation { return m.routerTree.List() }

func (m *Mux) SetResponseEncoder(name protoreflect.FullName, f func(ctx *fiber.Ctx, msg proto.Message) error) {
	m.opts.responseInterceptors[name] = f
}

func (m *Mux) SetRequestDecoder(name protoreflect.FullName, f func(ctx *fiber.Ctx, msg proto.Message) error) {
	m.opts.requestInterceptors[name] = f
}

func (m *Mux) MatchOperation(method string, path string) (r result.Result[*MatchOperation]) {
	restTarget, err := m.routerTree.Match(method, path)
	if err != nil {
		return r.WithErr(errors.Wrapf(err, "path not found, method=%s path=%s", method, path))
	}

	return r.WithVal(restTarget)
}

func (m *Mux) GetOperationByName(name string) *GrpcMethod {
	act := m.opts.customOperationNames[name]
	if act == nil {
		return nil
	}

	return handleOperation(act)
}

func (m *Mux) GetOperation(operation string) *GrpcMethod {
	var opt = m.opts.handlers[operation]
	if opt == nil {
		return nil
	}

	return handleOperation(opt)
}

func (m *Mux) Handler(ctx *fiber.Ctx) error {
	matchOperation, err := m.routerTree.Match(ctx.Method(), string(ctx.Request().URI().Path()))
	if err != nil {
		return errors.WrapCaller(err)
	}

	values := make(url.Values)
	for _, v := range matchOperation.Vars {
		values.Set(strings.Join(v.Fields, "."), v.Value)
	}

	for k, v := range ctx.Queries() {
		values.Set(k, v)
	}

	mth := m.opts.handlers[matchOperation.Operation]
	if mth == nil {
		return errors.Format("grpc method not found, method=%s", matchOperation.Operation)
	}

	md := metadata.MD{}
	for k, v := range ctx.GetReqHeaders() {
		md.Append(k, v...)
	}

	stream := &streamHTTP{
		handler: ctx,
		ctx:     metadata.NewIncomingContext(ctx.Context(), md),
		method:  mth,
		params:  values,
		path:    matchOperation,
	}

	var in = mth.inputType.New().Interface()
	err = stream.RecvMsg(in)
	if err != nil {
		return errors.WrapCaller(err)
	}

	var out = mth.outputType.New().Interface()
	var header metadata.MD
	var trailer metadata.MD
	err = m.Invoke(stream.ctx, mth.grpcFullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		return errors.WrapCaller(err)
	}

	var hh = make(metadata.MD)
	for k, v := range header {
		hh.Set(k, v...)
	}

	for k, v := range trailer {
		hh.Set(k, v...)
	}

	for k, v := range hh {
		v = lo.Filter(v, func(item string, index int) bool { return item != "" })
		if len(v) == 0 {
			continue
		}

		ctx.Response().Header.Set(k, v[0])
	}

	ctx.Response().Header.Set(httputil.HeaderXRequestVersion, version.Version())
	ctx.Response().Header.Set(httputil.HeaderXRequestOperation, matchOperation.Operation)
	ctx.Response().Header.SetContentTypeBytes(ctx.Request().Header.ContentType())
	return errors.WrapCaller(stream.SendMsg(out))
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
	adaptor.FiberHandler(m.Handler).ServeHTTP(writer, request)
}

func NewMux(opts ...MuxOption) *Mux {
	muxOpts := defaultMuxOptions
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

	for k := range muxOpts.codecs {
		muxOpts.contentTypeOffers = append(muxOpts.contentTypeOffers, k)
	}
	sort.Strings(muxOpts.contentTypeOffers)

	// Ensure compressors are set.
	if muxOpts.compressors == nil {
		muxOpts.compressors = make(map[string]Compressor)
	}

	for k, v := range defaultCompressors {
		if _, ok := muxOpts.compressors[k]; !ok {
			muxOpts.compressors[k] = v
		}
	}

	for k := range muxOpts.codecs {
		muxOpts.encodingTypeOffers = append(muxOpts.encodingTypeOffers, k)
	}
	sort.Strings(muxOpts.encodingTypeOffers)

	mux := &Mux{
		opts:        &muxOpts,
		localClient: new(inprocgrpc.Channel),
		routerTree:  routertree.New(),
	}

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
func (m *Mux) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	assert.If(generic.IsNil(ss), "ss params is nil")

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

func (m *Mux) registerService(gsd *grpc.ServiceDesc, ss interface{}, cli grpc.ClientConnInterface) error {
	d, err := m.opts.files.FindDescriptorByName(protoreflect.FullName(gsd.ServiceName))
	if err != nil {
		return errors.WrapCaller(err)
	}

	sd, ok := d.(protoreflect.ServiceDescriptor)
	if !ok {
		return errors.Format("invalid httpPathRule descriptor %T", d)
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

		assert.Exit(handlerHttpRoute(getExtensionHTTP(methodDesc), func(mth string, path string, reqBody, rspBody string) error {
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

		assert.Exit(handlerHttpRoute(getExtensionHTTP(methodDesc), func(mth string, path string, reqBody, rspBody string) error {
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
