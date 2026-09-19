package gateway

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pubgo/lava/v2/pkg/gateway/internal"
)

// ensureResponseCompression is allowed to run late — inside the SendStreamWriter
// callback, after fasthttp owns the response head. By then the pooled
// *fasthttp.RequestCtx may already be reset for the next connection, so the
// inbound compression headers must come from the handler-goroutine snapshot.
func TestStreamHTTP_CompressionNegotiationUsesRequestSnapshot(t *testing.T) {
	app := fiber.New()
	fctx := &fasthttp.RequestCtx{}
	ctx := app.AcquireCtx(fctx)
	defer app.ReleaseCtx(ctx)

	s := &streamHTTP{
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
	s.reqCT = "application/grpc-web+proto"
	s.handler.Request().Header.Set("Grpc-Accept-Encoding", "gzip")
	s.snapshotRequestEncoding()

	// What fasthttp does once the response is done with this request.
	s.handler.Request().Header.Del("Grpc-Accept-Encoding")

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
		t.Fatalf("want a gzip frame from the snapshot, got flags=%b", raw[0])
	}
	n := binary.BigEndian.Uint32(raw[1:5])
	plain, err := decompressMessage(s.compressorByName("gzip"), raw[5:5+int(n)])
	if err != nil {
		t.Fatal(err)
	}
	if got := string(plain); !strings.Contains(got, "payload-") {
		t.Fatalf("payload lost in frame: %q", got)
	}
}

// Inbound frames are decoded with the request's grpc-encoding, which RecvMsg
// reads before the response is committed but the same hazard applies.
func TestStreamHTTP_RequestEncodingUsesRequestSnapshot(t *testing.T) {
	s := testStreamHTTPWithGzip(t)
	s.handler.Request().Header.Set("Grpc-Encoding", "gzip")
	s.snapshotRequestEncoding()
	s.handler.Request().Header.Del("Grpc-Encoding")

	if got := s.requestEncoding(); got != "gzip" {
		t.Fatalf("requestEncoding=%q want gzip", got)
	}
}
