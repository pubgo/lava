package gateway

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/pubgo/lava/v2/core/encoding/protojson"
	"github.com/pubgo/lava/v2/pkg/gateway/gatewayutils"
	"github.com/pubgo/lava/v2/pkg/gateway/routertree"
)

type streamHTTP struct {
	method     *methodWrapper
	path       *routertree.MatchOperation
	handler    fiber.Ctx
	ctx        context.Context
	header     metadata.MD
	trailer    metadata.MD
	params     url.Values
	sentHeader bool
	// responseStream indicates this stream writes multiple response messages.
	// For JSON transport we emit NDJSON (one JSON object per line).
	responseStream bool
	writer     io.Writer // optional custom writer
}

var _ grpc.ServerStream = (*streamHTTP)(nil)

func (s *streamHTTP) SetHeader(md metadata.MD) error {
	if s.sentHeader {
		return errors.WrapStack(fmt.Errorf("already sent headers"))
	}
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamHTTP) SendHeader(md metadata.MD) error {
	if s.sentHeader {
		return errors.WrapCaller(fmt.Errorf("already sent headers"))
	}
	s.header = metadata.Join(s.header, md)
	s.sentHeader = true

	for k, v := range s.header {
		if len(v) == 0 {
			continue
		}
		// HTTP header 通常只支持单个值，取第一个值
		s.handler.Response().Header.Set(k, v[0])
	}

	return nil
}

func (s *streamHTTP) SetTrailer(md metadata.MD) {
	if s.trailer == nil {
		s.trailer = make(metadata.MD)
	}
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamHTTP) Context() context.Context {
	return NewContextWithServerTransportStream(s.ctx, s, s.method.grpcFullMethod)
}

func (s *streamHTTP) SendMsg(m any) error {
	if funk.IsNil(m) {
		return errors.New("stream http send msg got nil")
	}

	reply, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream http send proto msg got unknown type message")
	}

	if fRsp, ok := s.handler.Response().BodyWriter().(http.Flusher); ok {
		defer fRsp.Flush()
	}

	cur := reply.ProtoReflect()
	for _, fd := range getRspBodyDesc(s.path) {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	reqName := msg.ProtoReflect().Descriptor().FullName()
	rspInterceptor := s.method.srv.opts.responseInterceptors[reqName]
	if rspInterceptor != nil {
		return errors.Wrapf(rspInterceptor(s.handler, msg), "failed to do rsp interceptor response data by %s", reqName)
	}

	ct := string(s.handler.Request().Header.ContentType())
	isGRPC := strings.HasPrefix(ct, "application/grpc")

	var b []byte
	var err error
	if isGRPC {
		b, err = proto.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "failed to marshal response by protobuf")
		}
		// Add gRPC frame header: compression(0) + message type(0) + length
		frame := make([]byte, 5+len(b))
		binary.BigEndian.PutUint32(frame[1:5], uint32(len(b)))
		copy(frame[5:], b)
		b = frame
	} else {
		b, err = protojson.Default.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "failed to marshal response by protojson")
		}
	}

	if s.writer != nil {
		_, err = s.writer.Write(b)
	} else {
		_, err = s.handler.Write(b)
	}
	if err != nil {
		return errors.WrapCaller(err)
	}

	if !isGRPC && s.responseStream {
		if s.writer != nil {
			_, err = s.writer.Write([]byte("\n"))
		} else {
			_, err = s.handler.Write([]byte("\n"))
		}
		if err != nil {
			return errors.WrapCaller(err)
		}
	}

	if s.writer != nil {
		if flusher, ok := s.writer.(interface{ Flush() }); ok {
			flusher.Flush()
		}
	}

	return nil
}

func (s *streamHTTP) RecvMsg(m any) error {
	if funk.IsNil(m) {
		return errors.New("stream http recv msg got nil")
	}

	args, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream http recv proto msg got unknown type message")
	}

	method := s.handler.Method()
	hasBody := method == http.MethodPut || method == http.MethodPost || method == http.MethodPatch
	allowBody := hasBody || method == http.MethodDelete

	if allowBody {
		cur := args.ProtoReflect()
		for _, fd := range getReqBodyDesc(s.path) {
			cur = cur.Mutable(fd).Message()
		}
		msg := cur.Interface()

		reqName := msg.ProtoReflect().Descriptor().FullName()
		reqInterceptor := s.method.srv.opts.requestInterceptors[reqName]
		if reqInterceptor != nil {
			return errors.Wrapf(reqInterceptor(s.handler, msg), "failed to go req interceptor request data by %s", reqName)
		}

		ct := string(s.handler.Request().Header.ContentType())
		isGRPC := strings.HasPrefix(ct, "application/grpc")

		// PUT/POST/PATCH 必须有 body (gRPC 请求除外，因为需要先解析帧)
		if hasBody && !isGRPC && len(s.handler.Body()) == 0 {
			return errors.WrapCaller(fmt.Errorf("request body is nil, operation=%s", reqName))
		}

		if s.handler.Request().IsBodyStream() {
			reader := s.handler.Request().BodyStream()
			if isGRPC {
				// Read gRPC frame header: 1 byte flags + 4 bytes length
				header := make([]byte, 5)
				if _, err := io.ReadFull(reader, header); err != nil {
					return errors.WrapCaller(err)
				}
				length := binary.BigEndian.Uint32(header[1:5])
				data := make([]byte, length)
				if _, err := io.ReadFull(reader, data); err != nil {
					return errors.WrapCaller(err)
				}
				if err := proto.Unmarshal(data, msg); err != nil {
					return errors.Wrapf(err, "failed to unmarshal body by protobuf, msg=%#v", msg)
				}
			} else {
				var b json.RawMessage
				if err := json.NewDecoder(reader).Decode(&b); err != nil {
					return errors.WrapCaller(err)
				}

				if err := protojson.Default.Unmarshal(b, msg); err != nil {
					return errors.Wrapf(err, "failed to unmarshal body by proto-json, msg=%#v", msg)
				}
			}
		} else {
			body := s.handler.Body()
			if isGRPC {
				// gRPC frame: 1 byte flags + 4 bytes length + message
				if len(body) < 5 {
					return errors.New("invalid gRPC frame: too short")
				}
				length := binary.BigEndian.Uint32(body[1:5])
				if len(body) < int(5+length) {
					return errors.Errorf("invalid gRPC frame: expected %d bytes, got %d", 5+length, len(body))
				}
				data := body[5 : 5+length]
				if err := proto.Unmarshal(data, msg); err != nil {
					return errors.Wrapf(err, "failed to unmarshal body by protobuf, msg=%#v", msg)
				}
			} else if len(body) > 0 {
				if err := protojson.Default.Unmarshal(body, msg); err != nil {
					return errors.Wrapf(err, "failed to unmarshal body by proto-json, msg=%#v", msg)
				}
			}
		}
	}

	if len(s.params) > 0 {
		if err := gatewayutils.PopulateQueryParameters(args, s.params, gatewayutils.NewDoubleArray(nil)); err != nil {
			return errors.Wrapf(err, "failed to set query params, params=%v", s.params)
		}
	}

	return nil
}
