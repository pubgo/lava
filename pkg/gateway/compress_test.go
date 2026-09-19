package gateway

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pubgo/lava/v2/pkg/gateway/internal"
)

func testStreamHTTPWithGzip(t *testing.T) *streamHTTP {
	t.Helper()
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	t.Cleanup(func() { app.ReleaseCtx(ctx) })

	return &streamHTTP{
		handler: ctx,
		method: &methodWrapper{
			srv: &serviceWrapper{
				opts: &muxOptions{
					compressors: map[string]Compressor{
						"gzip":     &internal.CompressorGzip{},
						"identity": nil,
					},
				},
			},
		},
	}
}

func TestNegotiateResponseCompressor_PrefersGzip(t *testing.T) {
	s := testStreamHTTPWithGzip(t)
	c, name := s.negotiateResponseCompressor("identity, gzip;q=1.0")
	if c == nil || name != "gzip" {
		t.Fatalf("got (%v, %q)", c, name)
	}
	if got := s.supportedAcceptEncoding(); got != "gzip,identity" {
		t.Fatalf("accept-encoding=%q", got)
	}
}

func TestDecodeGRPCFramePayload_GzipRoundTrip(t *testing.T) {
	s := testStreamHTTPWithGzip(t)
	s.handler.Request().Header.Set("Grpc-Encoding", "gzip")
	s.snapshotRequestEncoding()

	msg := wrapperspb.String("hello-compression")
	raw, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	comp, err := compressMessage(s.compressorByName("gzip"), raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.decodeGRPCFramePayload(grpcFrameCompressed, comp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestDecodeGRPCFramePayload_CompressedWithoutEncoding(t *testing.T) {
	s := testStreamHTTPWithGzip(t)
	_, err := s.decodeGRPCFramePayload(grpcFrameCompressed, []byte("x"))
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Fatalf("err=%v", err)
	}
}

func TestSendMsg_CompressesWhenClientAcceptsGzip(t *testing.T) {
	s := testStreamHTTPWithGzip(t)
	s.handler.Request().Header.SetContentType("application/grpc-web+proto")
	s.handler.Request().Header.Set("Grpc-Accept-Encoding", "gzip")
	s.snapshotRequestEncoding()

	var buf bytes.Buffer
	s.writer = &buf
	s.path = &MatchOperation{}

	msg := wrapperspb.String(strings.Repeat("payload-", 32))
	if err := s.SendMsg(msg); err != nil {
		t.Fatal(err)
	}
	raw := buf.Bytes()
	if len(raw) < 5 {
		t.Fatalf("short frame: %d", len(raw))
	}
	if raw[0]&grpcFrameCompressed == 0 {
		t.Fatal("expected compressed response frame")
	}
	if got := string(s.handler.Response().Header.Peek("Grpc-Encoding")); got != "gzip" {
		t.Fatalf("Grpc-Encoding=%q", got)
	}
	if got := string(s.handler.Response().Header.Peek("Grpc-Accept-Encoding")); !strings.Contains(got, "gzip") {
		t.Fatalf("Grpc-Accept-Encoding=%q", got)
	}

	n := binary.BigEndian.Uint32(raw[1:5])
	payload := raw[5 : 5+n]
	plain, err := decompressMessage(s.compressorByName("gzip"), payload)
	if err != nil {
		t.Fatal(err)
	}
	out := &wrapperspb.StringValue{}
	if err := proto.Unmarshal(plain, out); err != nil {
		t.Fatal(err)
	}
	if out.GetValue() != msg.GetValue() {
		t.Fatalf("value=%q", out.GetValue())
	}
}

func TestWriteTrailer_SkipsEncodingHeaders(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	ctx.Response().Header.Set("Grpc-Status", "0")
	ctx.Response().Header.Set("Grpc-Encoding", "gzip")
	ctx.Response().Header.Set("Grpc-Accept-Encoding", "gzip,identity")

	var buf bytes.Buffer
	w := &fiberWebWriter{ctx: ctx, typ: grpcWeb, enc: "proto", resp: &buf}
	if err := w.writeTrailer(); err != nil {
		t.Fatal(err)
	}
	body := string(buf.Bytes()[5:])
	if strings.Contains(strings.ToLower(body), "grpc-encoding") {
		t.Fatalf("trailer should skip encoding headers: %q", body)
	}
	if !strings.Contains(strings.ToLower(body), "grpc-status: 0") {
		t.Fatalf("trailer=%q", body)
	}
}
