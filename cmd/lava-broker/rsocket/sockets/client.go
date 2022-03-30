package sockets

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kr/pretty"
	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/service/service_type"
)

var XErr = xerror.New("rs")

var _ rsocket.RSocket = (*Handler)(nil)

type Handler struct {
	Client bool

	socket rsocket.RSocket

	handlers map[string]*Server
	// ctxHandlers 从context中获取值
	ctxHandlers []func(ctx context.Context)

	UnaryInterceptor  grpc.UnaryServerInterceptor
	StreamInterceptor grpc.StreamServerInterceptor

	middlewares []service_type.Middleware

	metaCodec encoding.Codec
	dataCodec encoding.Codec
	L         *logging.Logger `name:"rs.handler"`
}

func (t *Handler) Init() {
	defer xerror.RespExit()
	xerror.Assert(t.socket == nil, "[socket] is nil")
	xerror.Assert(t.metaCodec == nil, "[metaCodec] is nil")
	xerror.Assert(t.dataCodec == nil, "[dataCodec] is nil")
	xerror.Assert(t.L == nil, "[L] is nil")
}

// RegisterService 注册服务描述和handler实现
func (t *Handler) RegisterService(desc grpc.ServiceDesc, srv interface{}) {
	xerror.AssertFn(srv == nil || desc.HandlerType == nil, func() string {
		pretty.Println(srv)
		return "[desc] or [desc.HandlerType] or [srv] is nil"
	})

	// 类型检查, 检查srv是否实现desc.HandlerType接口
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(srv)
	if !st.Implements(ht) {
		t.L.Sugar().Fatalf("rsocket: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}

	if t.handlers == nil {
		t.handlers = make(map[string]*Server)
	}

	for i := range desc.Methods {
		var h = desc.Methods[i]
		t.handlers[desc.ServiceName+"/"+h.MethodName] = &Server{
			srv:     srv,
			handler: h.Handler,
		}
	}

	for i := range desc.Streams {
		var h = desc.Streams[i]
		t.handlers[desc.ServiceName+"/"+h.StreamName] = &Server{
			srv:           srv,
			stream:        h.Handler,
			ServerStreams: h.ServerStreams,
			ClientStreams: h.ClientStreams,
		}
	}
}

func (t *Handler) decode(msg payload.Payload) func(val interface{}) error {
	return func(val interface{}) error {
		return t.dataCodec.Unmarshal(msg.Data(), val)
	}
}

func (t *Handler) handleCtx(msg payload.Payload) context.Context {
	return context.Background()
}

func (t *Handler) unaryHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return handler(ctx, req)
}

func (t *Handler) getReq(msg payload.Payload) (*Request, error) {
	var md, ok = msg.Metadata()
	if ok && md != nil {
		var request Request
		return &request, t.metaCodec.Unmarshal(md, &request)
	}
	return nil, XErr.New("metadata not found")
}

func (t *Handler) checkHandle(req *Request) error {
	var h = t.handlers[req.Method]
	if h == nil || h.handler == nil {
		return XErr.New(fmt.Sprintf("request method not found: %s", req.Method))
	}
	return nil
}

// FireAndForget 对应grpc的服务定义 rpc Func(Message) returns (google.protobuf.Empty)
func (t *Handler) FireAndForget(msg payload.Payload) {
	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return
	}

	if logutil.ErrRecord(t.L, t.checkHandle(req)) {
		return
	}

	var h = t.handlers[req.Method]
	_, err = h.handler(h.srv, t.handleCtx(msg), t.decode(msg), t.unaryHandler)
	logutil.ErrRecord(t.L, err)
}

func (t *Handler) MetadataPush(msg payload.Payload) {}

func (t *Handler) RequestResponse(msg payload.Payload) mono.Mono {
	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	err = t.checkHandle(req)
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	var h = t.handlers[req.Method]
	resp, err := h.handler(h.srv, t.handleCtx(msg), t.decode(msg), func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		return resp, err
	})
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	dt, err := t.dataCodec.Marshal(resp)
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	return mono.Just(payload.New(dt, nil))
}

func (t *Handler) RequestStream(msg payload.Payload) flux.Flux {
	//msg
	//解码 得到metadata和方法名字, header
	//t.handlers 获取对应的handler

	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return flux.Error(err)
	}

	err = t.checkHandle(req)
	if logutil.ErrRecord(t.L, err) {
		return flux.Error(err)
	}

	var in = make(chan *ErrPayload)
	var out = make(chan *ErrPayload)

	go func() {
		in <- NewErrPayload(payload.New(reqData, nil))
	}()

	return flux.Create(func(ctx context.Context, s flux.Sink) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var h = t.handlers[req.Method]

		go func() {
			err := h.stream(h.srv, &serverStream{ctx: ctx, in: in, out: out, metaCodec: t.metaCodec, dataCodec: t.dataCodec})
			if logutil.ErrRecord(t.L, err) {
				s.Error(err)
				cancel()
				return
			}
		}()

		for {
			select {
			case pp := <-out:
				s.Next(pp)
				pp.Err <- nil
			case <-ctx.Done():
				s.Complete()
				return
			}
		}
	})
}

func (t *Handler) RequestChannel(msg flux.Flux) flux.Flux {
	return flux.Clone(msg).SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		v, ok := s.Value()
		if !ok && v == nil {
			return flux.Error(fmt.Errorf("get metadata fail"))
		}

		req, err := t.getReq(v)
		if logutil.ErrRecord(t.L, err) {
			return flux.Error(err)
		}

		err = t.checkHandle(req)
		if logutil.ErrRecord(t.L, err) {
			return flux.Error(err)
		}

		var in = make(chan *ErrPayload)
		var out = make(chan *ErrPayload)
		var h = t.handlers[req.Method]

		return flux.Create(func(ctx context.Context, sink flux.Sink) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			var stream = &serverStream{ctx: ctx, in: in, out: out, metaCodec: t.metaCodec, dataCodec: t.dataCodec}
			go func() {
				if err := h.stream(h.srv, stream); err != nil {
					sink.Error(err)
					cancel()
				}
			}()

			f.DoOnNext(func(input payload.Payload) error {
				var pp = NewErrPayload(input)
				in <- pp
				return <-pp.Err
			}).DoOnError(func(e error) {
				cancel()
				logging.L().Error("err", logutil.ErrField(e)...)
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
