package sockets

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/errors"
)

var _ grpc.ClientConnInterface = (*Client)(nil)

type Client struct {
	// contentType 默认内容类型, 由应用设置
	contentType string
	socket      rsocket.RSocket

	// ctxHandlers 从context中获取值
	ctxHandlers []func(ctx context.Context)
}

// Invoke 请求响应模型实现
func (t *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	// 从ctx获取metadata
	// 从opts获取编码等
	// FireAndForget
	// RequestResponse

	// 根据获取的metadata做log和metric
	// data.Metadata()

	// 从 opts 获取codec
	// 对结果进行解码
	// 超时

	//"missing metadata"
	// TODO 获取ContentType
	// ctxHandlers

	var optCall = GetCallOptions(opts...)

	contentType := t.contentType
	if optCall.ContentSubtype != "" {
		contentType = optCall.ContentSubtype
	}

	var cdc = encoding.GetCodec(contentType)
	if optCall.Codec != nil {
		cdc = optCall.Codec
	}

	// 方法名字 内容类型 版本 链路 请求id 链路 请求方
	var req = &Request{
		Method:      method,
		ContentType: contentType,
	}

	reqMd, err := proto.Marshal(req)
	if err != nil {
		return xerror.Wrap(err)
	}

	data, err := cdc.Marshal(args)
	if err != nil {
		return xerror.Wrap(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	switch reply.(type) {
	// TODO reply 为*emptypb.Empty的情况
	case *emptypb.Empty:
		t.socket.FireAndForget(payload.New(data, reqMd))
		return nil
	default:
		var resp = t.socket.RequestResponse(payload.New(data, reqMd))
		pp, release, err := resp.BlockUnsafe(ctx)
		if err != nil {
			return xerror.Wrap(err)
		}

		defer release()

		var md, ok = pp.Metadata()
		if ok && md != nil {
			var respMd Response
			if err := proto.Unmarshal(md, &respMd); err != nil {
				return err
			}

			// code不是0, 服务端数据处理有问题
			if respMd.Code != 0 {
				return errors.New("server response", respMd.Code, respMd.Msg)
			}

			// TODO 处理 metadata
			//if respMd.Metadata.Header != nil {
			//	for k, v := range req.Metadata.Header {
			//		s.headers.Set(k, v)
			//	}
			//}
		}

		return cdc.Unmarshal(pp.Data(), reply)
	}
}

// NewStream 单项流和双向流实现
func (t *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// RequestStream
	// RequestChannel
	// 获取元数据

	var optCall = GetCallOptions(opts...)

	contentType := t.contentType
	if optCall.ContentSubtype != "" {
		contentType = optCall.ContentSubtype
	}

	var cdc = encoding.GetCodec(contentType)
	if optCall.Codec != nil {
		cdc = optCall.Codec
	}

	// 方法名字 内容类型 版本 链路 请求id 链路 请求方
	var req = &Request{
		Method:      method,
		ContentType: contentType,
	}

	reqMd, err := proto.Marshal(req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	ctx, cancel := context.WithCancel(ctx)

	var in = make(chan *ErrPayload)
	var out = make(chan *ErrPayload)

	go func() {
		t.socket.
			RequestChannel(flux.Create(func(ctx context.Context, s flux.Sink) {
				// 客户端输入数据
				for {
					select {
					case pp := <-in:
						s.Next(pp.Payload)
						pp.Err <- nil
					case <-ctx.Done():
						s.Complete()
						return
					}
				}
			})).
			DoOnNext(func(input payload.Payload) error {
				// 客户端接收数据
				var p = NewErrPayload(input)
				out <- p
				return <-p.Err
			}).
			DoOnError(func(err error) {
				cancel()
				logging.L().Error("err", logutil.ErrField(err)...)
			}).
			DoOnComplete(func() { cancel() }).
			DoFinally(func(s rx.SignalType) { cancel() }).Subscribe(ctx)
	}()

	return &clientStream{md: reqMd, cdc: cdc, in: in, out: out, ctx: ctx}, nil
}

type clientStream struct {
	reqMdOnce      sync.Once
	md             []byte
	cdc            encoding.Codec
	out            chan *ErrPayload
	in             chan *ErrPayload
	ctx            context.Context
	responseStream bool
	headers        metadata.MD
	trailers       metadata.MD
	socket         rsocket.RSocket
	sendClosed     bool
}

func (s *clientStream) Header() (metadata.MD, error) {
	return s.headers, nil
}

func (s *clientStream) Trailer() metadata.MD {
	return s.trailers
}

func (s *clientStream) CloseSend() error {
	s.sendClosed = true
	return nil
}

func (s *clientStream) Context() context.Context { return s.ctx }

// SendMsg 客户端发送数据到服务端
func (s *clientStream) SendMsg(m interface{}) error {
	if s.sendClosed {
		return ErrSendClosed
	}

	if m == nil {
		return errors.InvalidArgument("clientStream.SendMsg", "message to send is nil")
	}

	var data, err = s.cdc.Marshal(m)
	if err != nil {
		return xerror.Wrap(err, "clientStream.SendMsg Marshal error")
	}

	var md []byte
	// 客户端发送metadata, 只在第一次的时候发送
	// 服务端第一次接收数据解析metadata, 获取method等相关信息
	s.reqMdOnce.Do(func() { md = s.md })

	var p = NewErrPayload(payload.New(data, md))
	select {
	case s.in <- p:
		return <-p.Err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// RecvMsg 客户端从服务端接收数据
func (s *clientStream) RecvMsg(m interface{}) error {
	select {
	case resp := <-s.out:
		var md, ok = resp.Metadata()
		if ok && md != nil {
			var req Response
			if err := proto.Unmarshal(md, &req); err != nil {
				resp.Err <- err
				return err
			}

			// code不是0, 服务端数据处理有问题
			if req.Code != 0 {
				return errors.New("server response", req.Code, req.Msg)
			}

			// 处理 metadata
			//if req.Metadata.Header != nil {
			//	for k, v := range req.Metadata.Header {
			//		s.headers.Set(k, v)
			//	}
			//}
		}

		var err = s.cdc.Unmarshal(resp.Data(), m)

		// TODO 错误信息做处理
		// 把错误信息告诉服务方
		resp.Err <- err

		// 错误信息返回给业务处理者
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}
