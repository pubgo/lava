package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
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
// It writes headers directly to Fiber / fasthttp response headers.
type fiberWebWriter struct {
	ctx         fiber.Ctx
	fctx        *fasthttp.RequestCtx // used when Fiber ctx may be pooled (server-stream)
	resp        io.Writer
	respCloser  io.Closer // flushes streaming base64 encoder for grpc-web-text
	flushWriter http.Flusher
	typ         string // grpcWeb or grpcWebText
	enc         string // proto or json
	wroteHeader bool
	wroteResp   bool
}

func newFiberWebWriter(ctx fiber.Ctx, typ, enc string) *fiberWebWriter {
	return newFiberWebWriterTo(ctx, ctx.RequestCtx(), typ, enc, ctx.Response().BodyWriter())
}

// newFiberWebWriterTo wires gRPC-Web framing onto an arbitrary writer (e.g. Fiber
// SendStreamWriter's bufio.Writer) so server-streams can flush per message.
func newFiberWebWriterTo(ctx fiber.Ctx, fctx *fasthttp.RequestCtx, typ, enc string, raw io.Writer) *fiberWebWriter {
	resp := raw
	var respCloser io.Closer
	if typ == grpcWebText {
		bw := newBase64ChunkWriter(raw)
		resp = bw
		respCloser = bw
	}
	return &fiberWebWriter{
		ctx:         ctx,
		fctx:        fctx,
		typ:         typ,
		enc:         enc,
		resp:        resp,
		respCloser:  respCloser,
		flushWriter: asHTTPFlusher(raw),
	}
}

type flushErrAdapter struct {
	f interface{ Flush() error }
}

func (a flushErrAdapter) Flush() { _ = a.f.Flush() }

func asHTTPFlusher(w io.Writer) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if f, ok := w.(interface{ Flush() error }); ok {
		return flushErrAdapter{f}
	}
	return nil
}

func (w *fiberWebWriter) setContentType(v string) {
	if w.fctx != nil {
		w.fctx.Response.Header.Set("Content-Type", v)
		return
	}
	if w.ctx != nil {
		w.ctx.Set("Content-Type", v)
	}
}

func (w *fiberWebWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.setContentType(w.typ + "+" + w.enc)
	}
	w.wroteResp = true
	return w.resp.Write(data)
}

func (w *fiberWebWriter) writeTrailer() error {
	tr := make(http.Header)
	if w.fctx != nil {
		for key, value := range w.fctx.Response.Header.All() {
			k := string(key)
			if !isGRPCWebTrailerHeader(k) {
				continue
			}
			tr.Set(k, string(value))
		}
	} else if w.ctx != nil {
		//lint:ignore SA1019 VisitAll is the only available API in this fasthttp version.
		for key, value := range w.ctx.Response().Header.All() {
			k := string(key)
			if !isGRPCWebTrailerHeader(k) {
				continue
			}
			tr.Set(k, string(value))
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

func isGRPCWebTrailerHeader(k string) bool {
	k = strings.ToLower(k)
	switch k {
	case "grpc-encoding", "grpc-accept-encoding", "grpc-timeout", "grpc-message-type":
		return false
	default:
		return strings.HasPrefix(k, "grpc-")
	}
}

// markErrorTrailer is kept for call-site clarity on the error path.
// flushWithTrailer always emits a trailer frame.
func (w *fiberWebWriter) markErrorTrailer() {}

// ensureTrailer is kept for call-site clarity on the success path.
// flushWithTrailer always emits a trailer frame.
func (w *fiberWebWriter) ensureTrailer() {}

func (w *fiberWebWriter) flushWithTrailer() {
	// gRPC-Web clients always expect a trailer frame (success defaults to grpc-status=0).
	if !w.wroteHeader {
		w.wroteHeader = true
		w.setContentType(w.typ + "+" + w.enc)
	}
	if err := w.writeTrailer(); err != nil {
		return
	}
	// Must Close the streaming base64 encoder so residual bits and padding are flushed.
	// Per-Write Encode() with padding would concatenate into invalid base64 for clients.
	if w.respCloser != nil {
		_ = w.respCloser.Close()
		w.respCloser = nil
	}
	w.Flush()
}

func (w *fiberWebWriter) Flush() {
	if w.flushWriter != nil {
		w.flushWriter.Flush()
	}
}

// base64ChunkWriter streams binary frames as one continuous base64 body.
// Do not Encode each Write independently: padded chunks concatenated are not
// valid base64 and break browser atob / StdEncoding.DecodeString.
type base64ChunkWriter struct {
	enc io.WriteCloser
}

func newBase64ChunkWriter(w io.Writer) *base64ChunkWriter {
	return &base64ChunkWriter{enc: base64.NewEncoder(base64.StdEncoding, w)}
}

func (b *base64ChunkWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return b.enc.Write(p)
}

func (b *base64ChunkWriter) Close() error {
	if b.enc == nil {
		return nil
	}
	err := b.enc.Close()
	b.enc = nil
	return err
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
