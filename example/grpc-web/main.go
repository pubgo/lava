package main

import (
	"github.com/pubgo/lug/example/grpc_entry/handler"
	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type codec struct{}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (c *codec) Name() string {
	return "json"
}

func init() {
	encoding.RegisterCodec(&codec{})
}

func main() {
	grpcServer := grpc.NewServer()
	hello.RegisterTestApiServer(grpcServer, handler.NewTestAPIHandler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		fmt.Println(r.Header)

		var dd, err = ioutil.ReadAll(r.Body)
		xerror.Panic(err)
		fmt.Println("input", string(dd), "end")

		r.Body = grpcDataWrapper(dd)
		grpcServer.ServeHTTP(newGrpcWebResponse(w), hackIntoNormalGrpcRequest(r))
		return
	})

	http.ListenAndServe("127.0.0.1:8900", nil)
}

func grpcDataWrapper(data ...[]byte) io.ReadCloser {
	writer := new(bytes.Buffer)
	for _, msgBytes := range data {
		grpcPreamble := []byte{0, 0, 0, 0, 0}
		binary.BigEndian.PutUint32(grpcPreamble[1:], uint32(len(msgBytes)))
		writer.Write(grpcPreamble)
		writer.Write(msgBytes)
	}
	return ioutil.NopCloser(writer)
}

// 把http1参数转化为http2参数
func hackIntoNormalGrpcRequest(req *http.Request) *http.Request {
	// Hack, this should be a shallow copy, but let's see if this works
	req.ProtoMajor = 2
	req.ProtoMinor = 0

	req.Method = http.MethodPost

	req.Header.Set("content-type", "application/grpc+json")

	// Remove content-length header since it represents http1.1 payload size, not the sum of the h2
	// DATA frame payload lengths. https://http2.github.io/http2-spec/#malformed This effectively
	// switches to chunked encoding which is the default for h2
	req.Header.Del("content-length")

	// header处理

	return req
}

// grpcWebResponse implements http.ResponseWriter.
type grpcWebResponse struct {
	w http.ResponseWriter
}

func newGrpcWebResponse(resp http.ResponseWriter) *grpcWebResponse {
	g := &grpcWebResponse{w: resp}
	return g
}

func (w *grpcWebResponse) Header() http.Header {
	return w.w.Header()
}

func (w *grpcWebResponse) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *grpcWebResponse) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func (w *grpcWebResponse) Flush() {
	flushWriter(w.w)
}

func flushWriter(w http.ResponseWriter) {
	f, ok := w.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}
