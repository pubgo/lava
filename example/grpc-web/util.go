package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/goccy/go-json"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

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
	req.Header.Del("Content-Length")

	return req
}

const (
	headerContentLength  = "Content-Length"
	headerGRPCMessage    = "Grpc-Message"
	headerGRPCStatusCode = "Grpc-Status"
	headerUseInsecure    = "Grpc-Insecure"
)

//func handleGRPCResponse(resp http.ResponseWriter) http.ResponseWriter {
//	code := resp.Header().Get(headerGRPCStatusCode)
//	if code != "0" && code != "" {
//		buff := bytes.NewBuffer(nil)
//		grpcMessage := resp.Header().Get(headerGRPCMessage)
//		j, _ := json.Marshal(grpcMessage)
//		buff.WriteString(`{"error":` + string(j) + ` ,"code":` + code + `}`)
//
//		resp.Body = ioutil.NopCloser(buff)
//		resp.StatusCode = 500
//
//		return resp
//	}
//
//	prefix := make([]byte, 5)
//	_, _ = resp.Body.Read(prefix)
//
//	resp.Header.Del(headerContentLength)
//
//	return resp
//
//}

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

	fmt.Println(b)

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

// Transport struct for intercepting grpc+json requests
type Transport struct {
	HTTPClient    *http.Client
	H2Client      *http.Client
	H2NoTLSClient *http.Client
}

/*
	NewProxy returns a configured reverse proxy
	to handle grpc+json requests
*/
func NewProxy() *httputil.ReverseProxy {
	h2NoTLSClient := &http.Client{
		// Skip TLS dial
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		},
		Timeout: defaultClientTimeout,
	}

	h2Client := &http.Client{
		Transport: &http2.Transport{},
		Timeout:   defaultClientTimeout,
	}

	client := &http.Client{
		Timeout: defaultClientTimeout,
	}

	t := &Transport{
		HTTPClient:    client,
		H2Client:      h2Client,
		H2NoTLSClient: h2NoTLSClient,
	}

	u := url.URL{}
	p := httputil.NewSingleHostReverseProxy(&u)
	p.Director = t.director
	p.Transport = t

	return p
}

func (t Transport) director(r *http.Request) {}

/*
  RoundTrip handles processing the incoming request
  and outgoing response for grpc+json detection
*/
func (t Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	isGRPC := false
	if isJSONGRPC(r) {
		if r.Method != http.MethodPost {
			buff := bytes.NewBufferString("HTTP method must be POST")
			resp := &http.Response{
				StatusCode: 502,
				Body:       ioutil.NopCloser(buff),
			}
			return resp, nil
		}
		isGRPC = true
		r = modifyRequestToJSONgRPC(r)
	}

	client := t.HTTPClient
	if isGRPC {
		if r.Header.Get(headerUseInsecure) != "" {
			client = t.H2NoTLSClient
		} else {
			client = t.H2Client
		}
	}

	// clear requestURI, set in call to director
	r.RequestURI = ""

	log.Printf("proxying request url=[%s] isJSONGRPC=[%t]\n", r.URL.String(), isGRPC)

	resp, err := client.Do(r)
	if err != nil {
		log.Printf("unable to do request err=[%s]", err)

		buff := bytes.NewBuffer(nil)
		buff.WriteString(err.Error())
		resp = &http.Response{
			StatusCode: 502,
			Body:       ioutil.NopCloser(buff),
		}

		return resp, nil
	}

	if isGRPC {
		return handleGRPCResponse(resp)
	}

	return resp, err
}

const (
	// header to detect if it is a grpc+json request
	contentTypeGRPCJSON = "application/grpc+json"

	grpcNoCompression byte = 0x00

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
