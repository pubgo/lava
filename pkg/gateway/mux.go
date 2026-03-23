package gateway

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
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
	"github.com/pubgo/lava/v2/pkg/httputil"
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
	localClient *inprocgrpc.Channel
	opts        *muxOptions
	routerTree  *routertree.RouteTree
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
	// Check if this is a gRPC Web request
	ct := string(ctx.Request().Header.ContentType())
	if typ, enc, ok := isWebRequestFromContentType(ct, ctx.Method()); ok {
		// TODO: Check for websocket request and upgrade.
		if strings.EqualFold(ctx.Get("Upgrade"), "websocket") {
			return fiber.NewError(fiber.StatusInternalServerError, "unimplemented websocket support")
		}

		// Modify request for gRPC Web
		ctx.Request().Header.SetContentType(grpcBase + "+" + enc)
		if typ == grpcWebText {
			// gRPC-Web-Text (Base64) 解码处理
			// 策略：
			// 1. 如果是 Stream 模式 (Fasthttp BodyStream != nil)，则包裹 Stream 进行流式解码。
			// 2. 如果是 Buffer 模式 (Body 已经在内存中)，则直接对 Body 进行解码并回写。

			inputStream := ctx.Request().BodyStream()
			if inputStream != nil {
				// 流式处理
				body := base64.NewDecoder(base64.StdEncoding, inputStream)
				rc := &readCloser{
					Reader: body,
					Closer: io.NopCloser(nil),
				}
				ctx.Request().SetBodyStream(rc, -1)
			} else {
				// 非流式处理，直接操作 Body 字节
				originBody := ctx.Body()
				if len(originBody) > 0 {
					// Base64 解码需要分配新内存，这在普通请求中是可接受的
					// 计算解码后长度
					dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(originBody)))
					n, err := base64.StdEncoding.Decode(dbuf, originBody)
					if err == nil {
						ctx.Request().SetBody(dbuf[:n])
					} else {
						// 如果解码失败，这里暂时无法中止 Handler，只能留给后续 Protobuf Unmarshal 报错
						// 但至少不能 Panic
						log.Err(err).
							Stack().
							Str("method", ctx.Method()).
							Str("path", string(ctx.Request().URI().Path())).
							Msg("base64 decode failed")
						return errors.Errorf("base64 decode failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
					}
				}
			}
		}

		// Create Fiber-specific web writer
		ww := newFiberWebWriter(ctx, typ, enc)

		// Continue with normal processing but capture the response
		matchOperation, err := m.routerTree.Match(ctx.Method(), string(ctx.Request().URI().Path()))
		if err != nil {
			log.Error().
				Str("method", ctx.Method()).
				Str("path", string(ctx.Request().URI().Path())).
				Msg("match operation failed")
			return errors.Errorf("match operation failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
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
			log.Error().
				Str("method", ctx.Method()).
				Str("path", string(ctx.Request().URI().Path())).
				Msg("method operation not found")
			return errors.Errorf("method operation not found, method=%s path=%s", matchOperation.Operation, ctx.Request().URI().Path())
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
			writer:  ww,
		}

		in := mth.inputType.New().Interface()
		err = stream.RecvMsg(in)
		if err != nil {
			log.Error().
				Str("method", ctx.Method()).
				Str("path", string(ctx.Request().URI().Path())).
				Msg("unmarshal request failed")
			return errors.Errorf("unmarshal request failed, method=%s", matchOperation.Operation)
		}

		ctx.Set(httputil.HeaderXRequestVersion, version.Version())
		ctx.Set(httputil.HeaderXRequestOperation, matchOperation.Operation)

		err = m.invokeWithStream(stream, in)
		if err != nil {
			log.Error().
				Str("method", ctx.Method()).
				Str("path", string(ctx.Request().URI().Path())).
				Msg("invoke failed")
			return errors.Errorf("invoke failed, method=%s", matchOperation.Operation)
		}
		ww.flushWithTrailer()
		return nil
	}

	matchOperation, err := m.routerTree.Match(ctx.Method(), string(ctx.Request().URI().Path()))
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("match operation failed")
		return errors.Errorf("match operation failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
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
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("method operation not found")
		return errors.Errorf("method operation not found, method=%s", matchOperation.Operation)
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

	in := mth.inputType.New().Interface()
	err = stream.RecvMsg(in)
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("unmarshal request failed")
		return errors.Errorf("unmarshal request failed, method=%s", matchOperation.Operation)
	}
	err = m.invokeWithStream(stream, in)
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("invoke failed")
		return errors.WrapCaller(err)
	}

	ctx.Response().Header.Set(httputil.HeaderXRequestVersion, version.Version())
	ctx.Response().Header.Set(httputil.HeaderXRequestOperation, matchOperation.Operation)
	ctx.Response().Header.SetContentTypeBytes(ctx.Request().Header.ContentType())
	return nil
}

func (m *Mux) invokeWithStream(stream *streamHTTP, in any) error {
	mth := stream.method
	if mth == nil {
		return errors.New("method wrapper is nil")
	}

	if mth.grpcStreamDesc != nil {
		return m.invokeResponseStream(stream, in)
	}

	out := mth.outputType.New().Interface()
	var header metadata.MD
	var trailer metadata.MD
	if err := m.Invoke(stream.ctx, mth.grpcFullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer)); err != nil {
		return err
	}

	applyResponseMetadata(stream.handler, header)
	applyResponseMetadata(stream.handler, trailer)

	return stream.SendMsg(out)
}

func (m *Mux) invokeResponseStream(stream *streamHTTP, in any) error {
	mth := stream.method
	if mth == nil || mth.grpcStreamDesc == nil {
		return errors.New("stream method descriptor is nil")
	}

	if !mth.grpcStreamDesc.ServerStreams {
		return errors.Errorf("unsupported stream mode: %s is not server-streaming", mth.grpcFullMethod)
	}
	if mth.grpcStreamDesc.ClientStreams {
		return errors.Errorf("unsupported stream mode: %s has client-streaming", mth.grpcFullMethod)
	}

	stream.responseStream = true

	clientStream, err := m.NewStream(stream.ctx, mth.grpcStreamDesc, mth.grpcFullMethod)
	if err != nil {
		return errors.WrapCaller(err)
	}

	if err = clientStream.SendMsg(in); err != nil {
		return errors.WrapCaller(err)
	}
	if err = clientStream.CloseSend(); err != nil {
		return errors.WrapCaller(err)
	}

	headerSent := false

	for {
		out := mth.outputType.New().Interface()
		err = clientStream.RecvMsg(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.WrapCaller(err)
		}

		if !headerSent {
			if header, headerErr := clientStream.Header(); headerErr == nil {
				if sendErr := stream.SendHeader(header); sendErr != nil {
					return errors.WrapCaller(sendErr)
				}
			}
			headerSent = true
		}

		if err = stream.SendMsg(out); err != nil {
			return errors.WrapCaller(err)
		}
	}

	if !headerSent {
		if header, headerErr := clientStream.Header(); headerErr == nil {
			if sendErr := stream.SendHeader(header); sendErr != nil {
				return errors.WrapCaller(sendErr)
			}
		}
	}

	stream.SetTrailer(clientStream.Trailer())
	applyResponseMetadata(stream.handler, stream.trailer)

	return nil
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
	// Check if this is a gRPC Web request
	if _, _, ok := isWebRequest(request); ok {
		serveGRPCWeb(m, writer, request)
		return
	}
	// For non-gRPC Web requests, we need to adapt to Fiber
	// Since Handler now handles gRPC Web internally, we can just use adaptor
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
