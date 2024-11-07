package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/lava/pkg/gateway/internal/routertree"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type streamHTTP struct {
	method     *methodWrapper
	path       *routertree.MatchOperation
	handler    fiber.Ctx
	ctx        context.Context
	header     metadata.MD
	params     url.Values
	sentHeader bool
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
		for i := range v {
			s.handler.Response().Header.Set(k, v[i])
		}
	}

	return nil
}

func (s *streamHTTP) SetTrailer(md metadata.MD) {
	s.header = metadata.Join(s.header, md)
}

func (s *streamHTTP) Context() context.Context {
	return NewContextWithServerTransportStream(s.ctx, s, s.method.grpcFullMethod)
}

func (s *streamHTTP) SendMsg(m interface{}) error {
	if generic.IsNil(m) {
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
	for _, fd := range getReqBodyDesc(s.path) {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	reqName := msg.ProtoReflect().Descriptor().FullName()
	rspInterceptor := s.method.srv.opts.responseInterceptors[reqName]
	if rspInterceptor != nil {
		return errors.Wrapf(rspInterceptor(s.handler, msg), "failed to do rsp interceptor response data by %s", reqName)
	}

	b, err := protojson.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal response by proto-json")
	}

	_, err = s.handler.Write(b)
	return errors.WrapCaller(err)
}

func (s *streamHTTP) RecvMsg(m interface{}) error {
	if generic.IsNil(m) {
		return errors.New("stream http recv msg got nil")
	}

	args, ok := m.(proto.Message)
	if !ok {
		return errors.New("stream http recv proto msg got unknown type message")
	}

	var method = s.handler.Method()

	if method == http.MethodPut ||
		method == http.MethodPost ||
		method == http.MethodDelete ||
		method == http.MethodPatch {
		cur := args.ProtoReflect()
		for _, fd := range getRspBodyDesc(s.path) {
			cur = cur.Mutable(fd).Message()
		}
		msg := cur.Interface()

		reqName := msg.ProtoReflect().Descriptor().FullName()
		reqInterceptor := s.method.srv.opts.requestInterceptors[reqName]
		if reqInterceptor != nil {
			return errors.Wrapf(reqInterceptor(s.handler, msg), "failed to go req interceptor request data by %s", reqName)
		}

		if method == http.MethodPut ||
			method == http.MethodPost ||
			method == http.MethodPatch {
			if len(s.handler.Body()) == 0 {
				return errors.WrapCaller(fmt.Errorf("request body is nil, operation=%s", reqName))
			}
		}

		if s.handler.Request().IsBodyStream() {
			var b json.RawMessage
			if err := json.NewDecoder(s.handler.Request().BodyStream()).Decode(&b); err != nil {
				return errors.WrapCaller(err)
			}

			if err := protojson.Unmarshal(b, msg); err != nil {
				return errors.Wrapf(err, "failed to unmarshal body by proto-json, msg=%#v", msg)
			}
		} else {
			if body := s.handler.Body(); len(body) > 0 {
				if err := protojson.Unmarshal(body, msg); err != nil {
					return errors.Wrapf(err, "failed to unmarshal body by proto-json, msg=%#v", msg)
				}
			}
		}
	}

	if len(s.params) > 0 {
		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
			return errors.Wrapf(err, "failed to set query params, params=%v", s.params)
		}
	}

	return nil
}
