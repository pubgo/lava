package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const (
	grpcBase    = "application/grpc"
	grpcWeb     = "application/grpc-web"
	grpcWebText = "application/grpc-web-text"
)

// isWebRequest checks for gRPC Web headers.
func isWebRequest(r *http.Request) (typ, enc string, ok bool) {
	ct := r.Header.Get("Content-Type")
	return isWebRequestFromContentType(ct, r.Method)
}

// isWebRequestFromContentType checks for gRPC Web headers from content type string.
func isWebRequestFromContentType(ct, method string) (typ, enc string, ok bool) {
	if !strings.HasPrefix(ct, "application/grpc-web") || method != http.MethodPost {
		return "", "", false
	}
	typ, enc, ok = strings.Cut(ct, "+")
	if !ok {
		enc = "proto"
	}
	ok = typ == grpcWeb || typ == grpcWebText
	return typ, enc, ok
}

// fiberWebWriter is a gRPC Web writer specifically for Fiber framework.
// It writes headers directly to Fiber response headers.
type fiberWebWriter struct {
	ctx         fiber.Ctx
	resp        io.Writer
	flushWriter http.Flusher
	typ         string // grpcWeb or grpcWebText
	enc         string // proto or json
	wroteHeader bool
	wroteResp   bool
}

func newFiberWebWriter(ctx fiber.Ctx, typ, enc string) *fiberWebWriter {
	raw := ctx.Response().BodyWriter()
	resp := raw
	if typ == grpcWebText {
		resp = &base64ChunkWriter{w: resp}
	}
	var flusher http.Flusher
	if f, ok := raw.(http.Flusher); ok {
		flusher = f
	}
	return &fiberWebWriter{
		ctx:         ctx,
		typ:         typ,
		enc:         enc,
		resp:        resp,
		flushWriter: flusher,
	}
}

func (w *fiberWebWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		// Set Content-Type header directly on Fiber response
		w.ctx.Set("Content-Type", w.typ+"+"+w.enc)
	}
	w.wroteResp = true
	return w.resp.Write(data)
}

func (w *fiberWebWriter) writeTrailer() error {
	// Write trailers only if message has been sent.
	if !w.wroteResp {
		return nil
	}
	tr := make(http.Header)
	// Collect grpc-* headers for trailer
	//lint:ignore SA1019 VisitAll is the only available API in this fasthttp version.
	for key, value := range w.ctx.Response().Header.All() {
		k := string(key)
		if strings.HasPrefix(strings.ToLower(k), "grpc-") {
			tr[strings.ToLower(k)] = []string{string(value)}
		}
	}
	// Add default grpc-status if not present
	if tr.Get("grpc-status") == "" {
		tr.Set("grpc-status", "0")
	}
	var buf bytes.Buffer
	if err := tr.Write(&buf); err != nil {
		return err
	}
	head := []byte{1 << 7, 0, 0, 0, 0} // MSB=1 indicates this is a trailer data frame.
	binary.BigEndian.PutUint32(head[1:5], uint32(buf.Len()))
	if _, err := w.resp.Write(head); err != nil {
		return err
	}
	if _, err := w.resp.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (w *fiberWebWriter) flushWithTrailer() {
	// Write trailers only if message has been sent.
	if w.wroteHeader || w.wroteResp {
		if err := w.writeTrailer(); err != nil {
			return // nothing
		}
	}
	w.Flush()
}

func (w *fiberWebWriter) Flush() {
	if w.flushWriter != nil {
		w.flushWriter.Flush()
	}
}

type base64ChunkWriter struct {
	w io.Writer
}

func (b *base64ChunkWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	out := make([]byte, base64.StdEncoding.EncodedLen(len(p)))
	base64.StdEncoding.Encode(out, p)
	if _, err := b.w.Write(out); err != nil {
		return 0, err
	}
	return len(p), nil
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (rc *readCloser) Read(p []byte) (n int, err error) {
	if rc.Reader == nil {
		return 0, io.EOF
	}

	return rc.Reader.Read(p)
}

func (rc *readCloser) Close() error {
	if rc.Closer == nil {
		return nil
	}

	return rc.Closer.Close()
}
