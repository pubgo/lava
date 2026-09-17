package gateway

import (
	"bytes"
	"io"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	grpcEncodingIdentity = "identity"
	grpcFrameCompressed  = 0x01
)

func errUnsupportedGRPCEncoding(enc string) error {
	if enc == "" {
		return status.Error(codes.Unimplemented, "compressed gRPC frame requires grpc-encoding")
	}
	return status.Errorf(codes.Unimplemented, "unsupported grpc-encoding %q", enc)
}

func normalizeContentCoding(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (s *streamHTTP) compressorByName(name string) Compressor {
	name = normalizeContentCoding(name)
	if name == "" || name == grpcEncodingIdentity {
		return nil
	}
	if s.method == nil || s.method.srv == nil || s.method.srv.opts == nil {
		return nil
	}
	c := s.method.srv.opts.compressors[name]
	if c == nil {
		return nil
	}
	return c
}

func (s *streamHTTP) supportedAcceptEncoding() string {
	names := make([]string, 0, 2)
	if s.method != nil && s.method.srv != nil && s.method.srv.opts != nil {
		for name, c := range s.method.srv.opts.compressors {
			if c == nil || normalizeContentCoding(name) == grpcEncodingIdentity {
				continue
			}
			names = append(names, normalizeContentCoding(name))
		}
	}
	if len(names) == 0 {
		return grpcEncodingIdentity
	}
	sort.Strings(names)
	preferred := make([]string, 0, len(names)+1)
	rest := make([]string, 0, len(names))
	for _, n := range names {
		if n == "gzip" {
			preferred = append(preferred, "gzip")
			continue
		}
		rest = append(rest, n)
	}
	preferred = append(preferred, rest...)
	preferred = append(preferred, grpcEncodingIdentity)
	return strings.Join(preferred, ",")
}

func (s *streamHTTP) negotiateResponseCompressor(acceptEncoding string) (Compressor, string) {
	for _, part := range strings.Split(acceptEncoding, ",") {
		name := normalizeContentCoding(part)
		// Strip q-value if present: "gzip;q=1.0"
		if i := strings.IndexByte(name, ';'); i >= 0 {
			name = strings.TrimSpace(name[:i])
		}
		if name == "" || name == grpcEncodingIdentity || name == "*" {
			continue
		}
		if c := s.compressorByName(name); c != nil {
			return c, name
		}
	}
	return nil, ""
}

func (s *streamHTTP) ensureResponseCompression() {
	if s.compNegotiated {
		return
	}
	s.compNegotiated = true

	s.setResponseHeader("Grpc-Accept-Encoding", s.supportedAcceptEncoding())

	accept := s.requestHeaderPeek("Grpc-Accept-Encoding")
	if accept == "" {
		accept = s.requestHeaderPeek("grpc-accept-encoding")
	}
	if c, name := s.negotiateResponseCompressor(accept); c != nil {
		s.respCompressor = c
		s.respEncoding = name
		s.setResponseHeader("Grpc-Encoding", name)
	}
}

func (s *streamHTTP) requestHeaderPeek(key string) string {
	if s.fctx != nil {
		return string(s.fctx.Request.Header.Peek(key))
	}
	if s.handler != nil {
		return string(s.handler.Request().Header.Peek(key))
	}
	return ""
}

func compressMessage(c Compressor, data []byte) ([]byte, error) {
	if c == nil {
		return data, nil
	}
	var buf bytes.Buffer
	w, err := c.Compress(&buf)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompressMessage(c Compressor, data []byte) ([]byte, error) {
	if c == nil {
		return data, nil
	}
	r, err := c.Decompress(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if closer, ok := r.(io.Closer); ok {
		_ = closer.Close()
	}
	return out, nil
}

func (s *streamHTTP) requestEncoding() string {
	enc := s.requestHeaderPeek("Grpc-Encoding")
	if enc == "" {
		enc = s.requestHeaderPeek("grpc-encoding")
	}
	return normalizeContentCoding(enc)
}

func (s *streamHTTP) decodeGRPCFramePayload(flags byte, payload []byte) ([]byte, error) {
	if flags&grpcFrameCompressed == 0 {
		return payload, nil
	}
	enc := s.requestEncoding()
	c := s.compressorByName(enc)
	if c == nil {
		return nil, errUnsupportedGRPCEncoding(enc)
	}
	return decompressMessage(c, payload)
}
