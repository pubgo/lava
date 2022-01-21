package sockets

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

var _ rsocket.RSocket = (*handlerMap)(nil)

type Server struct {
	srv           interface{}
	handler       UnaryHandler
	stream        StreamHandler
	ServerStreams bool
	ClientStreams bool
}

type handlerMap struct {
	// contentType 默认内容类型, 由应用设置
	contentType string
	ctx         context.Context
	interceptor grpc.UnaryServerInterceptor
	handlers    map[string]*Server
}

// RegisterService 注册服务描述和handler实现
func (t *handlerMap) RegisterService(desc *grpc.ServiceDesc, srv interface{}) {
	// 类型检查, 检查srv是否实现desc.HandlerType接口
	if srv != nil {
		ht := reflect.TypeOf(desc.HandlerType).Elem()
		st := reflect.TypeOf(srv)
		if !st.Implements(ht) {
			logging.S().Fatalf("rsocket: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
		}
	}

	if t.handlers == nil {
		t.handlers = make(map[string]*Server)
	}

	for i := range desc.Methods {
		var h = desc.Methods[i]
		t.handlers[desc.ServiceName+"."+h.MethodName] = &Server{
			srv:     srv,
			handler: h.Handler,
		}
	}

	for i := range desc.Streams {
		var h = desc.Streams[i]
		t.handlers[desc.ServiceName+"."+h.StreamName] = &Server{
			srv:           srv,
			stream:        h.Handler,
			ServerStreams: h.ServerStreams,
			ClientStreams: h.ClientStreams,
		}
	}
}

func (t *handlerMap) FireAndForget(msg payload.Payload) {
	var method string
	var cdc = encoding.GetCodec(t.contentType)

	var md, ok = msg.Metadata()
	if ok && md != nil {
		var reqMd Request
		xerror.Panic(proto.Unmarshal(md, &reqMd))
		method = reqMd.Method
		cdc = encoding.GetCodec(reqMd.ContentType)
		return
	}

	var h = t.handlers[method]
	var _, err = h.handler(h.srv, nil, func(i interface{}) error { return cdc.Unmarshal(msg.Data(), i) }, nil)
	xerror.Panic(err)

	// 对应grpc的服务定义 rpc Func(Message) returns (google.protobuf.Empty)
}

func (t *handlerMap) MetadataPush(msg payload.Payload) {
	// TODO implement me
	panic("implement me")
}

func (t *handlerMap) RequestResponse(msg payload.Payload) mono.Mono {
	var reqMd, ok = msg.Metadata()
	if !ok || reqMd == nil {
		return mono.Error(fmt.Errorf("method not found"))
	}

	var req = &Request{}
	if err := proto.Unmarshal(reqMd, req); err != nil {
		return mono.Error(err)
	}

	var method = req.Method
	var handler = t.handlers[method]
	if handler == nil || handler.handler == nil {
		return mono.Error(fmt.Errorf("method not found"))
	}

	var contentType = t.contentType
	if req.ContentType != "" {
		contentType = req.ContentType
	}
	var cdc = encoding.GetCodec(contentType)

	var reqData = msg.Data()

	// 响应 header
	return mono.Create(func(ctx context.Context, sink mono.Sink) {
		// msg
		// 解码 得到metadata和方法名字, header
		// t.handlers 获取对应的handler, 得到handler和srv

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// TODO middleware 处理
		// TODO 获取 处理后的header

		var respMd = Response{}

		// 执行 handler
		resp, err := handler.handler(
			handler.srv, ctx, func(in interface{}) error { return cdc.Unmarshal(reqData, in) }, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				// 获取元数据
				return handler(ctx, req)
			})
		if err != nil {
			// TODO error 处理
			sink.Error(err)
			return
		}

		data, err := cdc.Marshal(resp)
		if err != nil {
			// TODO error 处理
			sink.Error(err)
			return
		}

		md, err := proto.Marshal(&respMd)
		if err != nil {
			sink.Error(err)
			return
		}

		// TODO header 处理
		sink.Success(payload.New(data, md))
	})
}

func (t *handlerMap) RequestStream(msg payload.Payload) flux.Flux {
	//msg
	//解码 得到metadata和方法名字, header
	//t.handlers 获取对应的handler

	var reqMd, ok = msg.Metadata()
	if !ok || reqMd == nil {
		return flux.Error(fmt.Errorf("method not found"))
	}

	var req = &Request{}
	if err := proto.Unmarshal(reqMd, req); err != nil {
		return flux.Error(err)
	}

	var method = req.Method
	var handler = t.handlers[method]
	if handler == nil || handler.handler == nil {
		return flux.Error(fmt.Errorf("method not found"))
	}

	var contentType = t.contentType
	if req.ContentType != "" {
		contentType = req.ContentType
	}
	var cdc = encoding.GetCodec(contentType)

	var reqData = msg.Data()

	var in = make(chan *ErrPayload)
	var out = make(chan *ErrPayload)

	go func() {
		in <- NewErrPayload(payload.New(reqData, nil))
	}()

	return flux.Create(func(ctx context.Context, s flux.Sink) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// TODO header放进去
		var stream = &serverStream{ctx: ctx, in: in, out: out, cdc: cdc}
		go func() {
			// handler 服务端handler
			if err := handler.stream(handler.srv, stream); err != nil {
				cancel()
				s.Error(err)
			}
		}()

		for {
			select {
			case pp := <-in:
				s.Next(pp)
				pp.Err <- nil
			case <-ctx.Done():
				s.Complete()
				return
			}
		}
	})
}

func (t *handlerMap) RequestChannel(msg flux.Flux) flux.Flux {
	return flux.Clone(msg).SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		v, ok := s.Value()
		if !ok && v == nil {
			return flux.Error(fmt.Errorf("get metadata fail"))
		}

		reqMd, ok := v.Metadata()
		if !ok || reqMd == nil {
			return flux.Error(fmt.Errorf("method not found"))
		}

		var req = &Request{}
		if err := proto.Unmarshal(reqMd, req); err != nil {
			return flux.Error(err)
		}

		var method = req.Method
		var handler = t.handlers[method]
		if handler == nil || handler.handler == nil {
			return flux.Error(fmt.Errorf("method not found"))
		}

		var contentType = t.contentType
		if req.ContentType != "" {
			contentType = req.ContentType
		}
		var cdc = encoding.GetCodec(contentType)

		var in = make(chan *ErrPayload)
		var out = make(chan *ErrPayload)

		return flux.Create(func(ctx context.Context, sink flux.Sink) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var stream = &serverStream{ctx: ctx, in: in, out: out, cdc: cdc}
			go func() {
				if err := handler.stream(handler.srv, stream); err != nil {
					cancel()
					sink.Error(err)
				}
			}()

			f.DoOnNext(func(input payload.Payload) error {
				var pp = NewErrPayload(input)
				in <- pp
				return <-pp.Err
			}).DoOnError(func(e error) {
				cancel()
				logging.L().Error("err", logutil.WithErr(e)...)
			}).DoOnComplete(func() { cancel() }).Subscribe(ctx)

			for {
				select {
				case pp := <-out:
					sink.Next(pp.Payload)
					pp.Err <- nil
				case <-ctx.Done():
					sink.Complete()
					return
				}
			}
		})
	})
}

type serverStream struct {
	cdc         encoding.Codec
	in          chan *ErrPayload
	out         chan *ErrPayload
	ctx         context.Context
	onDone      context.CancelFunc
	headers     map[string]string
	sendHeaders map[string]string
	trailers    map[string]string
}

func (s *serverStream) SetHeader(md metadata.MD) error {
	return s.setHeader(md, false)
}

func (s *serverStream) SendHeader(md metadata.MD) error {
	return s.setHeader(md, true)
}

func (s *serverStream) setHeader(md metadata.MD, send bool) error {
	if s.headers == nil {
		s.headers = make(map[string]string)
	}
	for k, v := range md {
		s.headers[k] = strings.Join(v, ",")
	}

	if send {
		s.sendHeaders = s.headers
		s.headers = nil
	}
	return nil
}

func (s *serverStream) SetTrailer(md metadata.MD) {
	_ = s.TrySetTrailer(md) // must ignore return value
}

func (s *serverStream) TrySetTrailer(md metadata.MD) error {
	if s.trailers == nil {
		s.trailers = metadata.MD{}
	}
	for k, v := range md {
		s.trailers[k] = append(s.trailers[k], v...)
	}
	return nil
}

func (s *serverStream) Context() context.Context { return s.ctx }

// SendMsg 服务端发送数据到客户端
func (s *serverStream) SendMsg(m interface{}) error {
	var data, err = s.cdc.Marshal(m)
	if err != nil {
		return err
	}

	var respMd Response
	if s.sendHeaders != nil {
		// TODO 元数据发送到客户端
		respMd.Headers = s.sendHeaders
	}

	md, err := proto.Marshal(&respMd)
	if err != nil {
		return err
	}

	var pp = NewErrPayload(payload.New(data, md))
	s.out <- pp
	return <-pp.Err
}

// RecvMsg 服务端接收客户端数据
func (s *serverStream) RecvMsg(m interface{}) error {
	var pp = <-s.in

	var md, ok = pp.Metadata()
	if ok && md != nil {
		// TODO 元数据处理
		// 元数据调整header等
	}

	var err = s.cdc.Unmarshal(pp.Data(), m)
	// 把错误信息反馈给客户端
	pp.Err <- err
	return err
}
