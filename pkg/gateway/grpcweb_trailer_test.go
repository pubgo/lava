package gateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"io"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// closeSpy records whether an encoder was closed; flushSpy the same for flushes.
type closeSpy struct{ closed int }

func (c *closeSpy) Close() error { c.closed++; return nil }

type flushSpy struct{ flushed int }

func (f *flushSpy) Flush() { f.flushed++ }

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) { return 0, io.ErrClosedPipe }

func TestFiberWebWriter_TrailerFailureStillClosesEncoderAndFlushes(t *testing.T) {
	// A failed trailer write must not skip cleanup: the response is already lost,
	// but leaving the grpc-web-text base64 encoder unclosed abandons its state and
	// never pushes the buffered bytes to the response.
	closer := &closeSpy{}
	flusher := &flushSpy{}
	w := &fiberWebWriter{
		typ:              grpcWebText,
		enc:              "proto",
		resp:             failWriter{},
		respCloser:       closer,
		flushWriter:      flusher,
		wroteHeader:      true,
		headersCommitted: true,
	}

	w.flushWithTrailer()

	if closer.closed != 1 {
		t.Fatalf("encoder close calls=%d want 1 after a failed trailer write", closer.closed)
	}
	if flusher.flushed == 0 {
		t.Fatal("response must still be flushed after a failed trailer write")
	}
}

func TestFiberWebWriter_TrailerFrameUsesLowercaseFieldNames(t *testing.T) {
	var buf bytes.Buffer
	w := &fiberWebWriter{
		typ:              grpcWeb,
		enc:              "proto",
		resp:             &buf,
		headersCommitted: true,
	}
	w.addTrailers(metadata.Pairs("x-stream", "live", "grpc-message", "done"))
	w.markErrorTrailer(codes.NotFound, "missing")
	w.flushWithTrailer()

	raw := buf.Bytes()
	if len(raw) < 5 || raw[0]&0x80 == 0 {
		t.Fatalf("want a trailer frame, got %q", raw)
	}
	n := int(binary.BigEndian.Uint32(raw[1:5]))
	block := string(raw[5 : 5+n])

	seen := make(map[string]int)
	for _, line := range strings.Split(strings.TrimRight(block, "\r\n"), "\r\n") {
		k, _, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("malformed trailer line %q in %q", line, block)
		}
		if k != strings.TrimSpace(k) || k != strings.ToLower(k) {
			t.Fatalf("trailer field name %q must be lowercase, grpc-web parses trailer keys case-sensitively: %q", k, block)
		}
		seen[k]++
	}
	for _, k := range []string{"grpc-status", "grpc-message"} {
		if seen[k] != 1 {
			t.Fatalf("%s emitted %d times, want 1: %q", k, seen[k], block)
		}
	}
	// The default-status fallback must not clobber the error code.
	if !strings.Contains(block, "grpc-status: 5") {
		t.Fatalf("want grpc-status: 5 (NotFound), got %q", block)
	}
	if seen["x-stream"] != 1 {
		t.Fatalf("x-stream emitted %d times, want 1: %q", seen["x-stream"], block)
	}
}

func TestApplyGRPCWebMetadata_AllowsGRPCStatus(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	applyGRPCWebMetadata(ctx, metadata.MD{
		"grpc-status":  {"0"},
		"grpc-message": {"ok"},
		"x-custom":     {"v"},
	})

	if got := string(ctx.Response().Header.Peek("grpc-status")); got != "0" {
		t.Fatalf("grpc-status=%q", got)
	}
	if got := string(ctx.Response().Header.Peek("grpc-message")); got != "ok" {
		t.Fatalf("grpc-message=%q", got)
	}
	if got := string(ctx.Response().Header.Peek("x-custom")); got != "v" {
		t.Fatalf("x-custom=%q", got)
	}
}

func TestFiberWebWriter_SuccessTrailerDefaultsStatusZero(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	var buf bytes.Buffer
	w := &fiberWebWriter{
		ctx:  ctx,
		typ:  grpcWeb,
		enc:  "proto",
		resp: &buf,
	}
	w.flushWithTrailer()

	if buf.Len() < 5 {
		t.Fatalf("expected trailer frame, got %d bytes", buf.Len())
	}
	if buf.Bytes()[0]&0x80 == 0 {
		t.Fatal("MSB should mark trailer frame")
	}
	n := binary.BigEndian.Uint32(buf.Bytes()[1:5])
	body := string(buf.Bytes()[5 : 5+n])
	if !strings.Contains(strings.ToLower(body), "grpc-status: 0") {
		t.Fatalf("trailer body=%q, want grpc-status: 0", body)
	}
	if ct := string(ctx.Response().Header.Peek("Content-Type")); ct != "application/grpc-web+proto" {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestFiberWebWriter_SuccessTrailerKeepsAppliedStatus(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	applyGRPCWebMetadata(ctx, metadata.Pairs("grpc-status", "0", "grpc-message", "done"))

	var buf bytes.Buffer
	w := &fiberWebWriter{
		ctx:  ctx,
		typ:  grpcWeb,
		enc:  "proto",
		resp: &buf,
	}
	// Simulate a response body already written.
	if _, err := w.Write([]byte{0, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	w.flushWithTrailer()

	raw := buf.Bytes()
	// Skip the 5-byte data frame header + empty payload, then read trailer.
	if len(raw) < 10 {
		t.Fatalf("short response: %d", len(raw))
	}
	off := 5 + int(binary.BigEndian.Uint32(raw[1:5]))
	if off+5 > len(raw) || raw[off]&0x80 == 0 {
		t.Fatalf("missing trailer frame at %d: %v", off, raw)
	}
	n := binary.BigEndian.Uint32(raw[off+1 : off+5])
	body := string(raw[off+5 : off+5+int(n)])
	if !strings.Contains(strings.ToLower(body), "grpc-status: 0") {
		t.Fatalf("trailer=%q", body)
	}
	if !strings.Contains(strings.ToLower(body), "grpc-message: done") {
		t.Fatalf("trailer missing message: %q", body)
	}
}

func TestFiberWebWriter_MarkErrorTrailerOverridesStatus(t *testing.T) {
	var buf bytes.Buffer
	w := &fiberWebWriter{
		typ:  grpcWeb,
		enc:  "proto",
		resp: &buf,
	}
	w.markErrorTrailer(codes.InvalidArgument, "bad name")
	w.flushWithTrailer()

	if buf.Len() < 5 {
		t.Fatalf("expected trailer frame, got %d bytes", buf.Len())
	}
	n := binary.BigEndian.Uint32(buf.Bytes()[1:5])
	body := string(buf.Bytes()[5 : 5+n])
	if !strings.Contains(strings.ToLower(body), "grpc-status: 3") {
		t.Fatalf("trailer=%q, want grpc-status: 3", body)
	}
	if !strings.Contains(strings.ToLower(body), "grpc-message:") {
		t.Fatalf("trailer missing message: %q", body)
	}
}

func TestFiberWebWriter_AddTrailersInFrame(t *testing.T) {
	var buf bytes.Buffer
	w := &fiberWebWriter{
		typ:  grpcWeb,
		enc:  "proto",
		resp: &buf,
	}
	w.addTrailers(metadata.Pairs("x-demo-trailer", "stream-done"))
	w.flushWithTrailer()

	n := binary.BigEndian.Uint32(buf.Bytes()[1:5])
	body := string(buf.Bytes()[5 : 5+n])
	if !strings.Contains(strings.ToLower(body), "x-demo-trailer: stream-done") {
		t.Fatalf("trailer=%q, want x-demo-trailer", body)
	}
}

func TestFiberWebWriter_TextFormatStreamsValidBase64(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	var buf bytes.Buffer
	bw := newBase64ChunkWriter(&buf)
	w := &fiberWebWriter{
		ctx:        ctx,
		typ:        grpcWebText,
		enc:        "proto",
		resp:       bw,
		respCloser: bw,
	}
	// Data frame (empty payload) then trailer — multiple Writes must decode as one base64 blob.
	if _, err := w.Write([]byte{0, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	w.flushWithTrailer()

	raw, err := base64.StdEncoding.DecodeString(buf.String())
	if err != nil {
		t.Fatalf("response is not valid base64 (%q): %v", buf.String(), err)
	}
	if len(raw) < 10 {
		t.Fatalf("short decoded body: %d", len(raw))
	}
	if raw[0]&0x80 != 0 {
		t.Fatal("first frame should be data")
	}
	off := 5 + int(binary.BigEndian.Uint32(raw[1:5]))
	if off+5 > len(raw) || raw[off]&0x80 == 0 {
		t.Fatalf("missing trailer after decode: %v", raw)
	}
	if ct := string(ctx.Response().Header.Peek("Content-Type")); ct != "application/grpc-web-text+proto" {
		t.Fatalf("content-type=%q", ct)
	}
}

func TestBase64ChunkWriter_ConcatenatedWritesDecode(t *testing.T) {
	var buf bytes.Buffer
	w := newBase64ChunkWriter(&buf)
	// Intentionally write lengths that are not multiples of 3 so padding would
	// appear if each Write were encoded independently.
	if _, err := w.Write([]byte{1, 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte{3, 4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := base64.StdEncoding.DecodeString(buf.String())
	if err != nil {
		t.Fatalf("decode: %v (body=%q)", err, buf.String())
	}
	want := []byte{1, 2, 3, 4, 5, 6}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestWriteTrailer_UsesCanonicalHeaderKeys(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	// Error path sets Grpc-Status via fasthttp; trailer writer must not also emit 0.
	ctx.Response().Header.Set("Grpc-Status", "3")
	ctx.Response().Header.Set("Grpc-Message", "bad%20name")

	var buf bytes.Buffer
	w := &fiberWebWriter{ctx: ctx, typ: grpcWeb, enc: "proto", resp: &buf}
	if err := w.writeTrailer(); err != nil {
		t.Fatal(err)
	}
	body := string(buf.Bytes()[5:])
	if strings.Count(strings.ToLower(body), "grpc-status:") != 1 {
		t.Fatalf("want exactly one grpc-status line, got %q", body)
	}
	if !strings.Contains(strings.ToLower(body), "grpc-status: 3") {
		t.Fatalf("trailer=%q", body)
	}
	if strings.Contains(body, "grpc-status: 0") || strings.Contains(body, "Grpc-Status: 0") {
		t.Fatalf("default status 0 must not override error status: %q", body)
	}
}

func TestMuxRPCMiddleware_WrapsDispatch(t *testing.T) {
	var hits atomic.Int32
	mux := NewMux()
	mux.UseRPCMiddleware(func(ctx context.Context, op *Operation, next RPCHandler) (metadata.MD, metadata.MD, error) {
		hits.Add(1)
		h, tr, err := next(ctx)
		if h == nil {
			h = metadata.MD{}
		}
		h.Set("x-rpc-mw", "1")
		return h, tr, err
	})

	// Dispatch with nil frontend/op should still enter runRPC then fail inside dispatcher.
	_, _, err := mux.Dispatch(context.Background(), nil, &Operation{
		FullMethod: "/x.Y/Z",
		InputType:  (&emptypb.Empty{}).ProtoReflect().Type(),
		OutputType: (&emptypb.Empty{}).ProtoReflect().Type(),
	}, &emptypb.Empty{})
	if hits.Load() != 1 {
		t.Fatalf("rpc middleware hits=%d want 1 (err=%v)", hits.Load(), err)
	}
}
