package gateway

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/pkg/gateway/internal/routex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type muxOptions struct {
	types                 protoregistry.MessageTypeResolver
	files                 *protoregistry.Files
	unaryInterceptor      grpc.UnaryServerInterceptor
	streamInterceptor     grpc.StreamServerInterceptor
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
	handlers              map[string]*methodWrap
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
		handlers:              make(map[string]*methodWrap),
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
	cc    *inprocgrpc.Channel
	opts  *muxOptions
	route *routex.RouteTrie
}

func (m *Mux) GetRouteMethods() []*routex.RouteTarget {
	return m.route.GetRouteMethods()
}

func (m *Mux) SetResponseEncoder(name protoreflect.FullName, f func(ctx *fiber.Ctx, msg proto.Message) error) {
	m.opts.responseInterceptors[name] = f
}

func (m *Mux) SetRequestDecoder(name protoreflect.FullName, f func(ctx *fiber.Ctx, msg proto.Message) error) {
	m.opts.requestInterceptors[name] = f
}

func (m *Mux) Handler(ctx *fiber.Ctx) error {
	restTarget, restVars, _ := m.route.Match(string(ctx.Request().URI().Path()), ctx.Method())
	if restTarget == nil {
		return errors.Wrapf(fiber.ErrNotFound, "path=%s", string(ctx.Request().URI().Path()))
	}

	var values = make(url.Values)
	for _, v := range restVars {
		values.Set(v.Name, v.Value)
	}

	for k, v := range ctx.Queries() {
		values.Set(k, v)
	}

	mth := m.opts.handlers[restTarget.GrpcMethodName]
	if mth == nil {
		return errors.NewFmt("grpc method not found, method=%s", restTarget.GrpcMethodName)
	}

	md := metadata.New(nil)
	for k, v := range ctx.GetReqHeaders() {
		md.Append(k, v...)
	}

	rspCtx := metadata.NewIncomingContext(ctx.Context(), md)
	return errors.WrapCaller(mth.Handle(&streamHTTP{
		handler: ctx,
		ctx:     rspCtx,
		method:  mth,
		params:  values,
		path:    restTarget,
	}))
}

func (m *Mux) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return m.cc.Invoke(ctx, method, args, reply, opts...)
}

func (m *Mux) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return m.cc.NewStream(ctx, desc, method, opts...)
}

func (m *Mux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	adaptor.FiberHandler(m.Handler).ServeHTTP(writer, request)
}

func NewMux(opts ...MuxOption) *Mux {
	var muxOpts = defaultMuxOptions
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
		opts:  &muxOpts,
		cc:    new(inprocgrpc.Channel),
		route: routex.NewRouteTrie(),
	}

	return mux
}

func (m *Mux) SetUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) {
	m.opts.unaryInterceptor = interceptor
	m.cc.WithServerUnaryInterceptor(interceptor)
}

// SetStreamInterceptor configures the in-process channel to use the
// given server interceptor for streaming RPCs when dispatching.
func (m *Mux) SetStreamInterceptor(interceptor grpc.StreamServerInterceptor) {
	m.opts.streamInterceptor = interceptor
	m.cc.WithServerStreamInterceptor(interceptor)
}

// RegisterService satisfies grpc.ServiceRegistrar for generated service code hooks.
func (m *Mux) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	assert.If(generic.IsNil(ss), "ss params is nil")

	m.cc.RegisterService(sd, ss)

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		log.Fatal().Msgf("gateway: RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	if err := m.registerService(sd, ss); err != nil {
		log.Fatal().Err(err).Msgf("gateway: RegisterService error: %v", err)
	}
}

func (m *Mux) registerRouter(rule *methodWrap) {
	m.opts.handlers[rule.grpcMethodName] = rule
}

func (m *Mux) registerService(gsd *grpc.ServiceDesc, ss interface{}) error {
	d, err := m.opts.files.FindDescriptorByName(protoreflect.FullName(gsd.ServiceName))
	if err != nil {
		return errors.WrapCaller(err)
	}

	sd, ok := d.(protoreflect.ServiceDescriptor)
	if !ok {
		return errors.Format("invalid httpPathRule descriptor %T", d)
	}

	var srv = &serviceWrap{
		opts:        m.opts,
		srv:         ss,
		serviceDesc: gsd,
		servicePB:   sd,
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

		m.registerRouter(&methodWrap{
			srv:            srv,
			methodDesc:     grpcMth,
			grpcMethod:     methodDesc,
			grpcMethodName: grpcMethod,
		})
		assert.Must(m.route.AddRoute(grpcMethod, methodDesc))
	}

	for i := range gsd.Streams {
		grpcMth := &gsd.Streams[i]
		grpcMethod := "/" + gsd.ServiceName + "/" + grpcMth.StreamName
		assert.If(m.opts.handlers[grpcMethod] != nil, "grpc httpPathRule has existed")

		methodDesc := findMethodDesc(grpcMth.StreamName)

		m.registerRouter(&methodWrap{
			srv:            srv,
			streamDesc:     grpcMth,
			grpcMethod:     methodDesc,
			grpcMethodName: grpcMethod,
		})
		assert.Must(m.route.AddRoute(grpcMethod, methodDesc))
	}

	return nil
}
