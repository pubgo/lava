package main

import (
	"context"
	"github.com/pubgo/lug/example/grpc_entry/handler"
	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"time"

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
	hello.RegisterTransportServer(grpcServer, &trans{})
	fmt.Println(grpcServer.GetServiceInfo())

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
	w     http.ResponseWriter
	bytes *bytes.Buffer
}

func newGrpcWebResponse(resp http.ResponseWriter) *grpcWebResponse {
	g := &grpcWebResponse{w: resp, bytes: bytes.NewBuffer(nil)}
	return g
}

func (w *grpcWebResponse) Header() http.Header {
	return w.w.Header()
}

func (w *grpcWebResponse) Write(b []byte) (int, error) {
	var index, err = w.w.Write(b)

	fmt.Println(b)

	flushWriter(w.w)

	return index, err

	w.bytes.Write(b)

	fmt.Println(w.bytes.Bytes())

	//for {
	grpcPreamble := []byte{0, 0, 0, 0, 0}
	readCount, err := w.bytes.Read(grpcPreamble)
	//fmt.Println(err)
	//if err == io.EOF {
	//	return 0, err
	//}

	if readCount != 5 || err != nil {
		return -1, err
	}

	payloadLength := binary.BigEndian.Uint32(grpcPreamble[1:])
	payloadBytes := make([]byte, payloadLength)

	readCount, err = w.bytes.Read(payloadBytes)
	fmt.Println(payloadLength, readCount)
	if uint32(readCount) != payloadLength || err != nil {
		return -1, err
	}

	fmt.Println(string(payloadBytes))
	if grpcPreamble[0]&(1<<7) == (1 << 7) { // MSB signifies the trailer parser
		return w.w.Write(payloadBytes)
	} else {
		return w.w.Write(payloadBytes)
	}
	//}
}

func (w *grpcWebResponse) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

func (w *grpcWebResponse) Flush() {
	flushWriter(w.w)
}

func flushWriter(w http.ResponseWriter) {
	f, ok := w.(http.Flusher)
	fmt.Println(ok)

	if !ok {
		return
	}

	f.Flush()
}

var _ hello.TransportServer = (*trans)(nil)

type trans struct {
}

func (t *trans) TestStream(server hello.Transport_TestStreamServer) error {
	return nil
}

func (t *trans) TestStream1(server hello.Transport_TestStream1Server) error {
	_, _ = server.Recv()
	return server.SendAndClose(nil)
}

func (t *trans) TestStream2(message *hello.Message, server hello.Transport_TestStream2Server) error {
	message.Header["check"] = "ok"
	message.Header["ctx"] = fmt.Sprintf("%#v", server.Context())

	xerror.Exit(server.SetHeader(metadata.Pairs("a", "a1")))
	server.SetTrailer(metadata.Pairs("SetTrailer", "1"))
	for i := 0; i < 10; i++ {
		message.Header[fmt.Sprintf("index: %d", i)] = fmt.Sprintf("index: %d", i)
		if err := server.Send(message); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	return nil
}

func (t *trans) TestStream3(ctx context.Context, message *hello.Message) (*hello.Message, error) {
	message.Header["check"] = "ok"
	message.Header["ctx"] = fmt.Sprintf("%#v", ctx)
	q.Q(ctx)
	return message, nil
}
