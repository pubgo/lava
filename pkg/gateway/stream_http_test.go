package gateway

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"
)

// maxReadRecorder reports the largest buffer a reader was asked to fill, which
// shows whether the gateway allocated a client-sized buffer before validating
// the gRPC frame length prefix.
type maxReadRecorder struct {
	io.Reader
	maxRead int
}

func (m *maxReadRecorder) Read(p []byte) (int, error) {
	if len(p) > m.maxRead {
		m.maxRead = len(p)
	}
	return m.Reader.Read(p)
}

func TestStreamHTTP_RecvMsg_RejectsOversizedStreamedGRPCFrame(t *testing.T) {
	const grpcMaxMsgSize = 4 << 20 // core/registry.DefaultMaxMsgSize

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}

	tests := []struct {
		name   string
		length uint32
	}{
		{name: "one byte over the limit", length: grpcMaxMsgSize + 1},
		{name: "client claims max uint32", length: math.MaxUint32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			fctx := &fasthttp.RequestCtx{}
			ctx := app.AcquireCtx(fctx)
			defer app.ReleaseCtx(ctx)

			header := make([]byte, 5)
			binary.BigEndian.PutUint32(header[1:5], tt.length)

			rec := &maxReadRecorder{Reader: bytes.NewReader(header)}
			ctx.Request().Header.SetMethod("POST")
			ctx.Request().SetBodyStream(rec, -1)

			s := &streamHTTP{
				handler:   ctx,
				reqCT:     "application/grpc-web+proto",
				reqMethod: http.MethodPost,
				ctx:       context.Background(),
				method: &methodWrapper{
					srv:            &serviceWrapper{opts: NewMux().opts},
					inputType:      inType,
					grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true},
				},
			}

			err = s.RecvMsg(&emptypb.Empty{})
			if err == nil {
				t.Fatalf("RecvMsg accepted a %d byte frame", tt.length)
			}
			if got := status.Convert(err).Code(); got != codes.InvalidArgument {
				t.Fatalf("code=%s want InvalidArgument, err=%v", got, err)
			}
			if !strings.Contains(err.Error(), "too large") {
				t.Fatalf("err=%v want a frame-size rejection", err)
			}
			if rec.maxRead > 1<<20 {
				t.Fatalf("gateway allocated %d bytes for a %d byte frame before rejecting it",
					rec.maxRead, tt.length)
			}
		})
	}
}

func TestStreamHTTP_RecvMsg_AcceptsEmptyStreamedGRPCFrame(t *testing.T) {
	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("find input type: %v", err)
	}

	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	rec := &maxReadRecorder{Reader: bytes.NewReader([]byte{0, 0, 0, 0, 0})}
	ctx.Request().Header.SetMethod("POST")
	ctx.Request().SetBodyStream(rec, -1)

	s := &streamHTTP{
		handler:   ctx,
		reqCT:     "application/grpc-web+proto",
		reqMethod: http.MethodPost,
		ctx:       context.Background(),
		method: &methodWrapper{
			srv:            &serviceWrapper{opts: NewMux().opts},
			inputType:      inType,
			grpcStreamDesc: &grpc.StreamDesc{ServerStreams: true},
		},
	}

	if err = s.RecvMsg(&emptypb.Empty{}); err != nil {
		t.Fatalf("a valid zero-length frame must still be accepted: %v", err)
	}
}

func TestIsGRPCContentType_GrpcWebJSONAlias(t *testing.T) {
	if got := isGRPCContentType("application/grpc-web-json"); got {
		t.Fatal("grpc-web-json alias should be treated as JSON transport")
	}

	if got := isGRPCContentType("application/grpc-web-json; charset=utf-8"); got {
		t.Fatal("grpc-web-json alias with charset should be treated as JSON transport")
	}

	if got := isGRPCContentType("application/grpc+proto"); !got {
		t.Fatal("application/grpc+proto should be treated as gRPC transport")
	}
}

func TestStreamHTTP_SendHeader_Idempotent(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	s := &streamHTTP{handler: ctx}

	if err := s.SendHeader(metadata.Pairs("x-test", "v1")); err != nil {
		t.Fatalf("first SendHeader returned error: %v", err)
	}
	if err := s.SendHeader(metadata.Pairs("x-test", "v2")); err != nil {
		t.Fatalf("second SendHeader should be idempotent, got error: %v", err)
	}

	if got := string(ctx.Response().Header.Peek("x-test")); got != "v2" {
		t.Fatalf("header value mismatch: got=%q want=%q", got, "v2")
	}
}

func TestStreamHTTP_SetHeader_AfterSendHeader_NoError(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	s := &streamHTTP{handler: ctx}

	if err := s.SendHeader(metadata.Pairs("x-test", "v1")); err != nil {
		t.Fatalf("SendHeader returned error: %v", err)
	}
	if err := s.SetHeader(metadata.Pairs("x-extra", "ok")); err != nil {
		t.Fatalf("SetHeader after SendHeader should not fail, got error: %v", err)
	}

	if got := string(ctx.Response().Header.Peek("x-extra")); got != "ok" {
		t.Fatalf("header value mismatch: got=%q want=%q", got, "ok")
	}
}
