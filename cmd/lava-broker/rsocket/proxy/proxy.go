package sockets

import (
	"context"
	"fmt"
	"strings"

	"github.com/pubgo/lava/cmd/lava-broker/rs_manager"
	"github.com/pubgo/lava/cmd/lava-broker/rsocket/sockets"
	"google.golang.org/grpc/metadata"

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

func (t *Handler) getReq(msg payload.Payload) (*sockets.Request, error) {
	var md, ok = msg.Metadata()
	if ok && md != nil {
		var request sockets.Request
		return &request, t.metaCodec.Unmarshal(md, &request)
	}
	return nil, XErr.New("metadata not found")
}

// FireAndForget 对应grpc的服务定义 rpc Func(Message) returns (google.protobuf.Empty)
func (t *Handler) FireAndForget(msg payload.Payload) {
	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return
	}

	var h = rs_manager.GetService(req.Service)
	if h == nil {
		panic("service not found")
	}
	h.FireAndForget(msg)
}

func (t *Handler) MetadataPush(msg payload.Payload) {

}

func (t *Handler) RequestResponse(msg payload.Payload) mono.Mono {
	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	err = t.checkHandle(req)
	if logutil.ErrRecord(t.L, err) {
		return mono.Error(err)
	}

	var h = t.services[req.Method]
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
	//t.services 获取对应的handler

	req, err := t.getReq(msg)
	if logutil.ErrRecord(t.L, err) {
		return flux.Error(err)
	}

	err = t.checkHandle(req)
	if logutil.ErrRecord(t.L, err) {
		return flux.Error(err)
	}

	var in = make(chan *sockets.ErrPayload)
	var out = make(chan *sockets.ErrPayload)

	go func() {
		in <- sockets.NewErrPayload(payload.New(reqData, nil))
	}()

	return flux.Create(func(ctx context.Context, s flux.Sink) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var h = t.services[req.Method]

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
		var h = t.services[req.Method]

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

type Server struct {
	srv           interface{}
	handler       sockets.UnaryHandler
	stream        sockets.StreamHandler
	ServerStreams bool
	ClientStreams bool
}

type serverStream struct {
	metaCodec   encoding.Codec
	dataCodec   encoding.Codec
	in          chan *sockets.ErrPayload
	out         chan *sockets.ErrPayload
	ctx         context.Context
	onDone      context.CancelFunc
	headers     map[string]string
	sendHeaders map[string]string
	trailers    metadata.MD
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
	var data, err = s.dataCodec.Marshal(m)
	if err != nil {
		return err
	}

	var respMd sockets.Response
	if s.sendHeaders != nil {
		// TODO 元数据发送到客户端
		respMd.Headers = s.sendHeaders
	}

	md, err := s.metaCodec.Marshal(&respMd)
	if err != nil {
		return err
	}

	var pp = sockets.NewErrPayload(payload.New(data, md))
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

	var err = s.dataCodec.Unmarshal(pp.Data(), m)
	// 把错误信息反馈给客户端
	pp.Err <- err
	return err
}
