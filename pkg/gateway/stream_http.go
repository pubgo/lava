package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/pubgo/funk/log"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	_ "github.com/gofiber/contrib/websocket"
	_ "github.com/gorilla/websocket"
	_ "nhooyr.io/websocket"
)

type streamHTTP struct {
	opts           *muxOptions
	method         *httpPathRule
	ctx            *fiber.Ctx
	header         metadata.MD
	trailer        metadata.MD
	params         url.Values
	contentType    string
	accept         string
	acceptEncoding string
	recvCount      int
	sendCount      int
	sentHeader     bool
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

	setOutgoingHeader(h, s.header)
	// don't write the header code, wait for the body.
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
			method:       s.method.grpcMethodName,
		},
	)
}

func (s *streamHTTP) writeMsg(c Codec, b []byte, contentType string) (int, error) {
	count := s.sendCount
	if count == 0 {
		h := s.wHeader
		h.Set("Content-Type", contentType)
		if !s.sentHeader {
			if err := s.SendHeader(nil); err != nil {
				return count, err
			}
		}
	}
	s.sendCount += 1
	if s.method.desc.IsStreamingServer() {
		codec, ok := c.(StreamCodec)
		if !ok {
			return count, fmt.Errorf("codec %s does not support streaming", codec.Name())
		}
		_, err := codec.WriteNext(s.w, b)
		return count, err
	}
	return count, s.opts.writeAll(s.w, b)
}

func (s *streamHTTP) SendMsg(m interface{}) error {
	reply := m.(proto.Message)

	fRsp, ok := s.w.(http.Flusher)
	if ok {
		defer fRsp.Flush()
	}

	cur := reply.ProtoReflect()
	for _, fd := range s.method.rspBody {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	contentType := s.accept
	c, err := s.getCodec(contentType, cur)
	if err != nil {
		return err
	}

	bytes := bytesPool.Get().(*[]byte)
	b := (*bytes)[:0]
	defer func() {
		if cap(b) < s.opts.maxReceiveMessageSize {
			*bytes = b
			bytesPool.Put(bytes)
		}
	}()

	if cur.Descriptor().FullName() == "google.api.HttpBody" {
		fds := cur.Descriptor().Fields()
		fdContentType := fds.ByName(protoreflect.Name("content_type"))
		fdData := fds.ByName(protoreflect.Name("data"))
		pContentType := cur.Get(fdContentType)
		pData := cur.Get(fdData)

		b = append(b, pData.Bytes()...)
		contentType = pContentType.String()
	} else {
		var err error
		b, err = c.MarshalAppend(b, msg)
		if err != nil {
			return status.Errorf(codes.Internal, "%s: error while marshaling: %v", c.Name(), err)
		}
	}

	if _, err := s.writeMsg(c, b, contentType); err != nil {
		return err
	}

	if stats := s.opts.statsHandler; stats != nil {
		// TODO: raw payload stats.
		stats.HandleRPC(s.ctx, outPayload(false, m, b, b, time.Now()))
	}
	return nil
}

func (s *streamHTTP) readMsg(c Codec, b []byte) (int, []byte, error) {
	if s.ctx.Request().ConnectionClose() {
		return s.recvCount, nil, io.EOF
	}

	count := s.recvCount
	s.recvCount += 1
	if s.method.desc.IsStreamingClient() {
		codec, ok := c.(StreamCodec)
		if !ok {
			return count, nil, fmt.Errorf("codec %q does not support streaming", codec.Name())
		}
		b = append(b, s.rbuf...)
		b, n, err := codec.ReadNext(b, s.r, s.opts.maxReceiveMessageSize)
		if err == io.EOF {
			s.rEOF, err = true, nil
		}
		s.rbuf = append(s.rbuf[:0], b[n:]...)
		return count, b[:n], err
	}
	b, err := s.opts.readAll(b, s.r)
	if err == io.EOF {
		s.rEOF, err = true, nil
	}
	return count, b, err
}

func (s *streamHTTP) getCodec(mediaType string, cur protoreflect.Message) (Codec, error) {
	codecType := string(cur.Descriptor().FullName())
	if c, ok := s.opts.codecs[codecType]; ok {
		return c, nil
	}
	codecType = mediaType
	if c, ok := s.opts.codecs[codecType]; ok {
		return c, nil
	}
	return nil, status.Errorf(codes.Internal, "no codec registered for content-type %q", mediaType)
}

func (s *streamHTTP) decodeRequestArgs(args proto.Message) (int, error) {
	bytes := bytesPool.Get().(*[]byte)
	b := (*bytes)[:0]
	defer func() {
		if cap(b) < s.opts.maxReceiveMessageSize {
			*bytes = b
			bytesPool.Put(bytes)
		}
	}()

	cur := args.ProtoReflect()
	for _, fd := range s.method.reqBody {
		cur = cur.Mutable(fd).Message()
	}
	msg := cur.Interface()

	c, err := s.getCodec(s.contentType, cur)
	if err != nil {
		return -1, err
	}

	var (
		count int
	)

	count, b, err = s.readMsg(c, b)
	if err != nil {
		return count, err
	}

	if cur.Descriptor().FullName() == "google.api.HttpBody" {
		fds := cur.Descriptor().Fields()
		fdContentType := fds.ByName("content_type")
		fdData := fds.ByName("data")
		cur.Set(fdContentType, protoreflect.ValueOfString(s.contentType))

		cpy := make([]byte, len(b))
		copy(cpy, b)
		cur.Set(fdData, protoreflect.ValueOfBytes(cpy))
	} else {
		if err := c.Unmarshal(b, msg); err != nil {
			return count, status.Errorf(codes.Internal, "%s: error while unmarshaling: %v", c.Name(), err)
		}
	}
	return count, nil
}

func (s *streamHTTP) RecvMsg(m interface{}) error {
	args := m.(proto.Message)

	if len(s.params) > 0 {
		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
			log.Err(err).Msg("failed to set params")
		}
	}

	if s.method.hasReqBody {
		var err error
		count, err = s.decodeRequestArgs(args)
		if err != nil {
			return err
		}
	}

	return nil
}

type twirpError struct {
	Code    string            `json:"code"`
	Message string            `json:"msg"`
	Meta    map[string]string `json:"meta"`
}

func (m *Mux) encError(w http.ResponseWriter, r *http.Request, err error) {
	s, _ := status.FromError(err)
	if isTwirp := r.Header.Get("Twirp-Version") != ""; isTwirp {
		accept := "application/json"

		w.Header().Set("Content-Type", accept)
		w.WriteHeader(HTTPStatusCode(s.Code()))

		codeStr := strings.ToLower(code.Code_name[int32(s.Code())])

		terr := &twirpError{
			Code:    codeStr,
			Message: s.Message(),
		}
		b, err := json.Marshal(terr)
		if err != nil {
			panic(err) // ...
		}
		w.Write(b) //nolint
		return
	}

	accept := negotiateContentType(r.Header, m.opts.contentTypeOffers, "application/json")
	c := m.opts.codecs[accept]

	w.Header().Set("Content-Type", accept)
	w.WriteHeader(HTTPStatusCode(s.Code()))

	b, err := c.Marshal(s.Proto())
	if err != nil {
		panic(err) // ...
	}
	w.Write(b) //nolint
}
