package main

import (
	"context"
	"github.com/pubgo/lava/pkg/encoding"
	_ "github.com/pubgo/lava/pkg/encoding/bytes"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"testing"
)

func TestName(t *testing.T) {
	//resp, err := http.Post("http://localhost:8900/hello.Transport/TestStream2", "application/grpc+json", strings.NewReader(`{"header":{"hello":"ok"}}`))
	//go func() {
	//client := http.Client{
	//	// Skip TLS dial
	//	Transport: &http2.Transport{
	//		AllowHTTP: true,
	//		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
	//			return net.Dial(netw, addr)
	//		},
	//	},
	//}

	//resp, err := http.Post("http://localhost:8900/hello.TestApi/Version?input=error", "application/grpc+json", strings.NewReader(`{"header":{"hello":"ok"}}`))
	//xerror.Panic(err)
	//fmt.Println(resp.ContentLength)
	//fmt.Println(resp.Header)
	//io.Copy(os.Stdout, resp.Body)
	//}()

	//err := c.cc.Invoke(ctx, "/hello.TestApi/Version", in, out, opts...)

	conn, err := grpc.DialContext(context.Background(), "localhost:8900", grpc.WithInsecure(), grpc.WithBlock())
	//var conn, err = grpcc.NewDirect("localhost:8900", func(cfg *grpcc.Cfg) {})
	xerror.Panic(err)
	var data []byte
	xerror.Panic(conn.Invoke(
		context.Background(),
		"/hello.TestApi/Version", []byte(`{"input":"11111"}`), &data,
		grpc.ForceCodec(encoding.Get("bytes")),
	))
	select {}

	//fmt.Println(resp.Body)
}
