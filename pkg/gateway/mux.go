// Copyright 2021 Edward McFarlane. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/log"
	"google.golang.org/genproto/googleapis/api/serviceconfig"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type handlerFunc func(*muxOptions, grpc.ServerStream) error

type muxOptions struct {
	types                 protoregistry.MessageTypeResolver
	files                 *protoregistry.Files
	serviceConfig         *serviceconfig.Service
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
	app                   *fiber.App
	handlers              map[string]handlerFunc
}

// readAll reads from r until an error or EOF and returns the data it read.
func (o *muxOptions) readAll(b []byte, r io.Reader) ([]byte, error) {
	var total int64
	for {
		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		total += int64(n)
		if total > int64(o.maxReceiveMessageSize) {
			return nil, fmt.Errorf("max receive message size reached")
		}
		if err != nil {
			return b, err
		}
	}
}

func (o *muxOptions) writeAll(dst io.Writer, b []byte) error {
	if len(b) > o.maxSendMessageSize {
		return fmt.Errorf("max send message size reached")
	}
	n, err := dst.Write(b)
	if err == nil && n != len(b) {
		return io.ErrShortWrite
	}
	return err
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
		app:                   fiber.New(),
		handlers:              make(map[string]handlerFunc),
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

func UnaryServerInterceptorOption(interceptor grpc.UnaryServerInterceptor) MuxOption {
	return func(opts *muxOptions) { opts.unaryInterceptor = interceptor }
}

func StreamServerInterceptorOption(interceptor grpc.StreamServerInterceptor) MuxOption {
	return func(opts *muxOptions) { opts.streamInterceptor = interceptor }
}

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

type Mux struct {
	opts muxOptions
	mu   sync.Mutex
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

	return &Mux{
		opts: muxOpts,
	}
}

func (m *Mux) GetApp() *fiber.App { return m.opts.app }

// RegisterService satisfies grpc.ServiceRegistrar for generated service code hooks.
func (m *Mux) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	assert.If(generic.IsNil(ss), "ss params is nil")

	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		log.Fatal().Msgf("gateway: RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	if err := m.registerService(sd, ss); err != nil {
		log.Fatal().Err(err).Msgf("gateway: RegisterService error: %v", err)
	}
}

func (m *Mux) registerService(gsd *grpc.ServiceDesc, ss interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, err := m.opts.files.FindDescriptorByName(protoreflect.FullName(gsd.ServiceName))
	if err != nil {
		return errors.WrapCaller(err)
	}

	sd, ok := d.(protoreflect.ServiceDescriptor)
	if !ok {
		return errors.Format("invalid httpPathRule descriptor %T", d)
	}

	mds := sd.Methods()

	findMethodDesc := func(methodName string) (protoreflect.MethodDescriptor, error) {
		md := mds.ByName(protoreflect.Name(methodName))
		if md == nil {
			return nil, fmt.Errorf("missing httpPathRule descriptor for %v", methodName)
		}
		return md, nil
	}

	for i := range gsd.Methods {
		grpcMth := &gsd.Methods[i]
		methodDesc, err := findMethodDesc(grpcMth.MethodName)
		if err != nil {
			return errors.WrapCaller(err)
		}

		grpcMethod := fmt.Sprintf("/%s/%s", gsd.ServiceName, grpcMth.MethodName)
		assert.If(m.opts.handlers[grpcMethod] != nil, "grpc httpPathRule has existed")

		m.opts.handlers[grpcMethod] = func(opts *muxOptions, stream grpc.ServerStream) error {
			ctx := stream.Context()

			reply, err := grpcMth.Handler(ss, ctx, stream.RecvMsg, opts.unaryInterceptor)
			if err != nil {
				return errors.WrapCaller(err)
			}

			return errors.WrapCaller(stream.SendMsg(reply))
		}

		m.opts.app.Post(grpcMethod, handlerWrap(&httpPathRule{
			opts:           &m.opts,
			desc:           methodDesc,
			httpMethod:     http.MethodPost,
			httpPath:       grpcMethod,
			rawHttpPath:    grpcMethod,
			grpcMethodName: grpcMethod,
			vars:           make(map[string]string),
			hasReqBody:     true,
			hasRspBody:     true,
		}))

		for _, mth := range getMethod(getExtensionHTTP(methodDesc), methodDesc, grpcMethod) {
			m.opts.app.Add(mth.httpMethod, mth.httpPath, handlerWrap(mth))
		}
	}

	for i := range gsd.Streams {
		grpcMth := &gsd.Streams[i]
		grpcMethod := "/" + gsd.ServiceName + "/" + grpcMth.StreamName
		assert.If(m.opts.handlers[grpcMethod] != nil, "grpc httpPathRule has existed")

		methodDesc, err := findMethodDesc(grpcMth.StreamName)
		if err != nil {
			return err
		}

		m.opts.handlers[grpcMethod] = func(opts *muxOptions, stream grpc.ServerStream) error {
			info := &grpc.StreamServerInfo{
				FullMethod:     grpcMethod,
				IsClientStream: grpcMth.ClientStreams,
				IsServerStream: grpcMth.ServerStreams,
			}

			if opts.streamInterceptor != nil {
				return opts.streamInterceptor(stream.Context(), stream, info, grpcMth.Handler)
			} else {
				return grpcMth.Handler(ss, stream)
			}
		}

		m.opts.app.Post(grpcMethod, handlerWrap(&httpPathRule{
			opts:           &m.opts,
			desc:           methodDesc,
			httpMethod:     http.MethodPost,
			httpPath:       grpcMethod,
			rawHttpPath:    grpcMethod,
			grpcMethodName: grpcMethod,
			vars:           make(map[string]string),
			hasReqBody:     true,
			hasRspBody:     true,
		}))

		for _, mth := range getMethod(getExtensionHTTP(methodDesc), methodDesc, grpcMethod) {
			m.opts.app.Add(mth.httpMethod, mth.httpPath, handlerWrap(mth))
		}
	}

	return nil
}
