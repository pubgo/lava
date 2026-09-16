package gateway

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

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
	// recvDone is set after a successful RecvMsg so subsequent reads return EOF.
	// HTTP/gRPC-Web request bodies are single-shot for unary and server-stream.
	recvDone bool
	// responseStream indicates this stream writes multiple response messages.
	// For JSON transport we emit NDJSON (one JSON object per line).
	responseStream bool
	writer         io.Writer // optional custom writer
}

var _ grpc.ServerStream = (*streamHTTP)(nil)

func (s *streamHTTP) SetHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	if s.sentHeader {
		for k, v := range md {
			if len(v) == 0 {
				continue
			}
			s.handler.Response().Header.Set(k, v[0])
		}
	}
	return nil
}

func (s *streamHTTP) SendHeader(md metadata.MD) error {
	s.header = metadata.Join(s.header, md)
	if s.sentHeader {
		for k, v := range md {
			if len(v) == 0 {
				continue
			}
			s.handler.Response().Header.Set(k, v[0])
		}
		return nil
	}
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

func isGRPCContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	// Treat grpc-web-json alias as plain JSON transport for compatibility.
	if strings.HasPrefix(ct, "application/grpc-web-json") {
		return false
	}

	return strings.HasPrefix(ct, "application/grpc")
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
	isGRPC := isGRPCContentType(ct)
	codec := s.lookupCodec(ct)

	var b []byte
	var err error
	if isGRPC {
		b, err = codec.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "failed to marshal response by protobuf")
		}
		// Add gRPC frame header: compression(0) + message type(0) + length
		frame := make([]byte, 5+len(b))
		binary.BigEndian.PutUint32(frame[1:5], uint32(len(b)))
		copy(frame[5:], b)
		b = frame
	} else {
		b, err = codec.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "failed to marshal response by codec")
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
	if s.recvDone {
		return io.EOF
	}

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
		isGRPC := isGRPCContentType(ct)
		codec := s.lookupCodec(ct)

		// PUT/POST/PATCH 必须有 body (gRPC 请求除外，因为需要先解析帧)
		if hasBody && !isGRPC && len(s.handler.Body()) == 0 {
			return status.Errorf(codes.InvalidArgument, "request body is nil, operation=%s", reqName)
		}

		if s.handler.Request().IsBodyStream() {
			reader := s.handler.Request().BodyStream()
			if isGRPC {
				// Read gRPC frame header: 1 byte flags + 4 bytes length
				header := make([]byte, 5)
				if _, err := io.ReadFull(reader, header); err != nil {
					return status.Errorf(codes.InvalidArgument, "read grpc frame header: %v", err)
				}
				length := binary.BigEndian.Uint32(header[1:5])
				data := make([]byte, length)
				if _, err := io.ReadFull(reader, data); err != nil {
					return status.Errorf(codes.InvalidArgument, "read grpc frame body: %v", err)
				}
				if err := codec.Unmarshal(data, msg); err != nil {
					return status.Errorf(codes.InvalidArgument, "failed to unmarshal body by codec: %v", err)
				}
			} else {
				var b json.RawMessage
				if err := json.NewDecoder(reader).Decode(&b); err != nil {
					return status.Errorf(codes.InvalidArgument, "decode json body: %v", err)
				}

				if err := codec.Unmarshal(b, msg); err != nil {
					return status.Errorf(codes.InvalidArgument, "failed to unmarshal body by codec: %v", err)
				}
			}
		} else {
			body := s.handler.Body()
			if isGRPC {
				// gRPC frame: 1 byte flags + 4 bytes length + message
				if len(body) < 5 {
					return status.Error(codes.InvalidArgument, "invalid gRPC frame: too short")
				}
				length := binary.BigEndian.Uint32(body[1:5])
				if len(body) < int(5+length) {
					return status.Errorf(codes.InvalidArgument, "invalid gRPC frame: expected %d bytes, got %d", 5+length, len(body))
				}
				data := body[5 : 5+length]
				if err := codec.Unmarshal(data, msg); err != nil {
					return status.Errorf(codes.InvalidArgument, "failed to unmarshal body by codec: %v", err)
				}
			} else if len(body) > 0 {
				if err := codec.Unmarshal(body, msg); err != nil {
					return status.Errorf(codes.InvalidArgument, "failed to unmarshal body by codec: %v", err)
				}
			}
		}
	}

	if len(s.params) > 0 {
		if err := gatewayutils.PopulateQueryParameters(args, s.params, gatewayutils.NewDoubleArray(nil)); err != nil {
			return errors.Wrapf(err, "failed to set query params, params=%v", s.params)
		}
	}

	s.recvDone = true
	return nil
}

// lookupCodec resolves a Codec from the mux registry by Content-Type.
// Falls back to Protobuf for gRPC framed types and JSON otherwise.
func (s *streamHTTP) lookupCodec(contentType string) Codec {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if s.method != nil && s.method.srv != nil && s.method.srv.opts != nil {
		if c, ok := s.method.srv.opts.codecs[ct]; ok && c != nil {
			return c
		}
		// application/grpc+json → try json codec by name
		if _, enc, ok := strings.Cut(ct, "+"); ok {
			if c, ok := s.method.srv.opts.codecsByName[enc]; ok && c != nil {
				return c
			}
		}
	}
	if isGRPCContentType(ct) {
		return CodecProto{}
	}
	return CodecJSON{}
}
