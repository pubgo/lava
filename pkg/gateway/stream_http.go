package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/pkg/gateway/internal/routex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type streamHTTP struct {
	method     *methodWrap
	path       *routex.RouteTarget
	handler    *fiber.Ctx
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
	return grpc.NewContextWithServerTransportStream(
		s.ctx,
		&serverTransportStream{
			ServerStream: s,
			method:       s.method.grpcMethodName,
		},
	)
}

func (s *streamHTTP) SendMsg(m interface{}) error {
	reply := m.(proto.Message)

	fRsp, ok := s.handler.Response().BodyWriter().(http.Flusher)
	if ok {
		defer fRsp.Flush()
	}

	cur := reply.ProtoReflect()
	for _, fd := range s.path.ResponseBodyFields {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	var reqName = msg.ProtoReflect().Descriptor().FullName()
	handler := s.method.srv.opts.responseInterceptors[reqName]
	if handler != nil {
		return errors.Wrapf(handler(s.handler, msg), "failed to handler response data by %s", reqName)
	}

	b, err := protojson.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal response by protojson")
	}

	_, err = s.handler.Write(b)
	return errors.WrapCaller(err)
}

func (s *streamHTTP) RecvMsg(m interface{}) error {
	args := m.(proto.Message)

	if s.path.HttpMethod == http.MethodPut ||
		s.path.HttpMethod == http.MethodPost ||
		s.path.HttpMethod == http.MethodPatch {
		cur := args.ProtoReflect()
		for _, fd := range s.path.RequestBodyFields {
			cur = cur.Mutable(fd).Message()
		}
		msg := cur.Interface()

		var reqName = msg.ProtoReflect().Descriptor().FullName()
		handler := s.method.srv.opts.requestInterceptors[reqName]
		if handler != nil {
			return errors.Wrapf(handler(s.handler, msg), "failed to handler request data by %s", reqName)
		}

		if s.handler.Body() != nil && len(s.handler.Body()) != 0 {
			err := protojson.Unmarshal(s.handler.Body(), msg)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal body by protojson")
			}
		}
	}

	if s.params != nil && len(s.params) > 0 {
		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
			return errors.Wrapf(err, "failed to set query params")
		}
	}

	return nil
}
