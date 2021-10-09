package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
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
		req.Body = reqDataWrapper(strutil.ToBytes(req.URL.RawQuery))
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
	w http.ResponseWriter
}

func newGrpcWebResponse(resp http.ResponseWriter) *grpcWebResponse {
	return &grpcWebResponse{w: resp}
}

func (w *grpcWebResponse) Header() http.Header {
	return w.w.Header()
}

func (w *grpcWebResponse) Write(b []byte) (int, error) {
	if len(b) == 5 && (b[0]&(1<<7) == (1<<7) || b[0] == 0) {
		return 0, nil
	}

	var index, err = w.w.Write(b)

	flush(w.w)

	return index, err
}

func (w *grpcWebResponse) WriteHeader(code int) { w.w.WriteHeader(code) }
func (w *grpcWebResponse) Flush()               { flush(w.w) }

func flush(w http.ResponseWriter) {
	f, ok := w.(http.Flusher)
	if !ok {
		return
	}

	f.Flush()
}

//w.bytes.Write(b)
//
//grpcPreamble := []byte{0, 0, 0, 0, 0}
//readCount, err := w.bytes.Read(grpcPreamble)
//if err == io.EOF {
//	return 0, err
//}
//
//if readCount != 5 || err != nil {
//	return -1, err
//}
//
//payloadLength := binary.BigEndian.Uint32(grpcPreamble[1:])
//w.readCount = payloadLength
//
//if w.bytes.Len() < int(w.readCount) {
//	return 0, err
//}
//
//payloadBytes := make([]byte, payloadLength)
//readCount, err = w.bytes.Read(payloadBytes)
//if uint32(readCount) != payloadLength || err != nil {
//	return -1, err
//}
//
//if grpcPreamble[0]&(1<<7) == (1 << 7) { // MSB signifies the trailer parser
//	return w.w.Write(payloadBytes)
//} else {
//	return w.w.Write(payloadBytes)
//}

func GrpcCallFrom(ctx *fiber.Ctx) (context.Context, []grpc.CallOption) {
	//grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD)
	return nil, nil
}
