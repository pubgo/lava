package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
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

type webWriter struct {
	w           http.ResponseWriter
	resp        io.Writer
	seenHeaders map[string]bool
	typ         string // grpcWeb or grpcWebText
	enc         string // proto or json
	wroteHeader bool
	wroteResp   bool
}

func newWebWriter(w http.ResponseWriter, typ, enc string) *webWriter {
	var resp io.Writer = w
	if typ == grpcWebText {
		resp = base64.NewEncoder(base64.StdEncoding, resp)
	}
	return &webWriter{
		w:    w,
		typ:  typ,
		enc:  enc,
		resp: resp,
	}
}

// fiberWebWriter is a gRPC Web writer specifically for Fiber framework.
// Unlike webWriter which writes headers to body (for standard http.ResponseWriter),
// this writes headers directly to Fiber response headers.
type fiberWebWriter struct {
	ctx         fiber.Ctx
	resp        io.Writer
	typ         string // grpcWeb or grpcWebText
	enc         string // proto or json
	wroteHeader bool
	wroteResp   bool
}

func newFiberWebWriter(ctx fiber.Ctx, typ, enc string) *fiberWebWriter {
	resp := ctx.Response().BodyWriter()
	if typ == grpcWebText {
		resp = base64.NewEncoder(base64.StdEncoding, resp)
	}
	return &fiberWebWriter{
		ctx:  ctx,
		typ:  typ,
		enc:  enc,
		resp: resp,
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
	w.ctx.Response().Header.VisitAll(func(key, value []byte) {
		k := string(key)
		if strings.HasPrefix(strings.ToLower(k), "grpc-") {
			tr[strings.ToLower(k)] = []string{string(value)}
		}
	})
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
	if flusher, ok := w.resp.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *webWriter) Header() http.Header {
	return w.w.Header()
}

func (w *webWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		hdr := w.Header()
		hdr.Set("Content-Type", w.typ+"+"+w.enc) // override content-type
		for k, v := range hdr {
			if strings.HasPrefix(strings.ToLower(k), "grpc-") {
				continue
			}
			for _, val := range v {
				w.resp.Write([]byte(k))
				w.resp.Write([]byte(": "))
				w.resp.Write([]byte(val))
				w.resp.Write([]byte("\r\n"))
			}
		}
		w.resp.Write([]byte("\r\n"))
	}
	w.wroteResp = true
	return w.resp.Write(data)
}

func (w *webWriter) WriteHeader(statusCode int) {
	w.w.WriteHeader(statusCode)
}

func (w *webWriter) Flush() {
	if flusher, ok := w.resp.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *webWriter) writeTrailer() error {
	// Write trailers only if message has been sent.
	if !w.wroteResp {
		return nil
	}
	tr := make(http.Header)
	for k, v := range w.Header() {
		if strings.HasPrefix(strings.ToLower(k), "grpc-") {
			tr[strings.ToLower(k)] = v
		}
	}
	var buf bytes.Buffer
	if err := tr.Write(&buf); err != nil {
		return err
	}
	head := []byte{1 << 7, 0, 0, 0, 0} // MSB=1 indicates this is a trailer data frame.
	binary.BigEndian.PutUint32(head[1:5], uint32(buf.Len()))
	if _, err := w.Write(head); err != nil {
		return err
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func (w *webWriter) flushWithTrailer() {
	// Write trailers only if message has been sent.
	if w.wroteHeader || w.wroteResp {
		if err := w.writeTrailer(); err != nil {
			return // nothing
		}
	}
	w.Flush()
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

func serveGRPCWeb(m *Mux, w http.ResponseWriter, r *http.Request) {
	typ, enc, ok := isWebRequest(r)
	if !ok {
		msg := fmt.Sprintf("invalid gRPC-Web content type: %v", r.Header.Get("Content-Type"))
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// TODO: Check for websocket request and upgrade.
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "unimplemented websocket support", http.StatusInternalServerError)
		return
	}

	r.ProtoMajor = 2
	r.ProtoMinor = 0
	hdr := r.Header
	hdr.Del("Content-Length")
	hdr.Set("Content-Type", grpcBase+"+"+enc)
	if typ == grpcWebText {
		body := base64.NewDecoder(base64.StdEncoding, r.Body)
		r.Body = &readCloser{body, r.Body}
	}
	ww := newWebWriter(w, typ, enc)
	adaptor.FiberHandler(m.Handler).ServeHTTP(ww, r)
	ww.flushWithTrailer()
}
