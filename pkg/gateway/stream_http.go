package gateway

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/log"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type streamHTTP struct {
	opts       *muxOptions
	method     *httpPathRule
	ctx        *fiber.Ctx
	header     metadata.MD
	trailer    metadata.MD
	params     url.Values
	sentHeader bool
}

var _ grpc.ServerStream = (*streamHTTP)(nil)

func (s *streamHTTP) SetHeader(md metadata.MD) error {
	if s.sentHeader {
		return fmt.Errorf("already sent headers")
	}
	s.header = metadata.Join(s.header, md)
	return nil
}

func (s *streamHTTP) SendHeader(md metadata.MD) error {
	if s.sentHeader {
		return fmt.Errorf("already sent headers")
	}
	s.header = metadata.Join(s.header, md)
	s.sentHeader = true
	return nil
}

func (s *streamHTTP) SetTrailer(md metadata.MD) {
	s.trailer = metadata.Join(s.trailer, md)
}

func (s *streamHTTP) Context() context.Context {
	return grpc.NewContextWithServerTransportStream(
		s.ctx.Context(),
		&serverTransportStream{
			ServerStream: s,
			method:       s.method.GrpcMethodName,
		},
	)
}

func (s *streamHTTP) SendMsg(m interface{}) error {
	reply := m.(proto.Message)

	fRsp, ok := s.ctx.Response().BodyWriter().(http.Flusher)
	if ok {
		defer fRsp.Flush()
	}

	cur := reply.ProtoReflect()
	for _, fd := range s.method.rspBody {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	var reqName = msg.ProtoReflect().Descriptor().FullName()
	handler := s.opts.responseInterceptors[reqName]
	if handler != nil {
		return errors.Wrapf(handler(s.ctx, msg), "failed to handler response data by %s", reqName)
	}

	b, err := protojson.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal response by protojson")
	}

	_, err = s.ctx.Write(b)
	return err
}

func (s *streamHTTP) RecvMsg(m interface{}) error {
	args := m.(proto.Message)

	if s.method.HasReqBody {
		cur := args.ProtoReflect()
		for _, fd := range s.method.reqBody {
			cur = cur.Mutable(fd).Message()
		}
		msg := cur.Interface()

		var reqName = msg.ProtoReflect().Descriptor().FullName()
		handler := s.opts.requestInterceptors[reqName]
		if handler != nil {
			return errors.Wrapf(handler(s.ctx, msg), "failed to handler request data by %s", reqName)
		}

		err := protojson.Unmarshal(s.ctx.Body(), msg)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal body by protojson")
		}
	}

	if s.params != nil && len(s.params) > 0 {
		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
			log.Err(err).Msg("failed to set params")
		}
	}

	return nil
}
