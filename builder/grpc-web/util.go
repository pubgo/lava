package grpcWeb

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
)

func reqDataWrapper(data ...[]byte) io.ReadCloser {
	writer := new(bytes.Buffer)
	for _, msgBytes := range data {
		grpcPreamble := []byte{0, 0, 0, 0, 0}
		binary.BigEndian.PutUint32(grpcPreamble[1:], uint32(len(msgBytes)))
		writer.Write(grpcPreamble)
		writer.Write(msgBytes)
	}
	return ioutil.NopCloser(writer)
}

// req2GrpcRequest 把http1参数转化为http2参数
func req2GrpcRequest(req *http.Request) *http.Request {
	if req.Method == http.MethodGet {
		req.Header.Set("content-type", "application/grpc+uri")
		req.Body = reqDataWrapper(strutil.ToBytes(strings.TrimSpace(req.URL.RawQuery)))
	} else {
		req.Header.Set("content-type", "application/grpc+json")
		var dd, err = ioutil.ReadAll(req.Body)
		xerror.Panic(err)
		req.Body = reqDataWrapper(dd)
	}

	req.ProtoMajor = 2
	req.ProtoMinor = 0

	req.Method = http.MethodPost

	// Remove content-length header since it represents http1.1 payload size, not the sum of the h2
	// DATA frame payload lengths. https://http2.github.io/http2-spec/#malformed This effectively
	// switches to chunked encoding which is the default for h2
	req.Header.Del("content-length")

	return req
}

// grpcWebResponse implements http.ResponseWriter.
type grpcWebResponse struct {
	w   http.ResponseWriter
	buf *bytes.Buffer
}

func newGrpcWebResponse(resp http.ResponseWriter) *grpcWebResponse {
	return &grpcWebResponse{w: resp, buf: bytes.NewBuffer(nil)}
}

func (w *grpcWebResponse) Header() http.Header {
	w.w.Header().Set("content-type", "application/json")
	return w.w.Header()
}

func (w *grpcWebResponse) Write(b []byte) (int, error) { return w.buf.Write(b) }
func (w *grpcWebResponse) WriteHeader(code int)        { w.w.WriteHeader(code) }
func (w *grpcWebResponse) Flush() {
	grpcPreamble := []byte{0, 0, 0, 0, 0}
	readCount, err := w.buf.Read(grpcPreamble)
	if err == io.EOF {
		return
	}

	if readCount != 5 || err != nil {
		return
	}

	payloadLen := binary.BigEndian.Uint32(grpcPreamble[1:])
	if w.buf.Len() < int(payloadLen) {
		return
	}

	payloadBytes := make([]byte, payloadLen)
	readCount, err = w.buf.Read(payloadBytes)
	if uint32(readCount) != payloadLen || err != nil {
		return
	}

	w.w.Write(payloadBytes)

	flush(w.w)
}

func flush(w http.ResponseWriter) {
	f, ok := w.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}

const (
	// header to detect if it is a grpc+json request
	contentTypeGRPCJSON = "application/grpc+json"

	grpcNoCompression byte = 0x00

	headerContentLength  = "Content-Length"
	headerGRPCMessage    = "Grpc-Message"
	headerGRPCStatusCode = "Grpc-Status"
	headerUseInsecure    = "Grpc-Insecure"

	defaultClientTimeout = time.Second * 60
)

// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
func modifyRequestToJSONgRPC(r *http.Request) *http.Request {
	var body []byte
	// read body so we can add the grpc prefix
	if r.Body != nil {
		body, _ = ioutil.ReadAll(r.Body)
	}

	b := make([]byte, 0, len(body)+5)
	buff := bytes.NewBuffer(b)

	// grpc prefix is
	// 	1 byte: compression indicator
	// 	4 bytes: content length (excluding prefix)
	_ = buff.WriteByte(grpcNoCompression) // 0 or 1, indicates compressed payload

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(body)))

	_, _ = buff.Write(lenBytes)
	_, _ = buff.Write(body)

	// create new request
	req, _ := http.NewRequest(r.Method, r.URL.String(), buff)
	req.Header = r.Header

	// remove content length header
	req.Header.Del(headerContentLength)

	return req

}

func isJSONGRPC(r *http.Request) bool {
	h := r.Header.Get("Content-Type")

	if h == contentTypeGRPCJSON {
		return true
	}

	return false
}

func handleGRPCResponse(resp *http.Response) (*http.Response, error) {
	code := resp.Header.Get(headerGRPCStatusCode)
	if code != "0" && code != "" {
		buff := bytes.NewBuffer(nil)
		grpcMessage := resp.Header.Get(headerGRPCMessage)
		j, _ := json.Marshal(grpcMessage)
		buff.WriteString(`{"error":` + string(j) + ` ,"code":` + code + `}`)

		resp.Body = ioutil.NopCloser(buff)
		resp.StatusCode = 500

		return resp, nil
	}

	prefix := make([]byte, 5)
	_, _ = resp.Body.Read(prefix)

	resp.Header.Del(headerContentLength)

	return resp, nil

}
