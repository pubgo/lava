package main

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/pkg/typex"
	"testing"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	_ "github.com/pubgo/lava/encoding/bytes"
)

func TestName(t *testing.T) {
	//resp, err := http.Post("http://localhost:8900/hello.Transport/TestStream2", "application/grpc+json", strings.NewReader(`{"header":{"hello":"ok"}}`))
	//go func() {
	//client := http.Srv{
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
	//var conn, err = grpcc.NewDirect("localhost:8900", func(cfg *grpcc.Base) {})
	xerror.Panic(err)

	//var cli = hello.NewTestApiClient(conn)
	//o, err := cli.Version(context.Background(), &hello.TestReq{Input: "hello1"}, grpc.ForceCodec(encoding.GetCodec("json")))
	//xerror.Panic(err)
	//fmt.Printf("%v", o)

	// 可以通过sidecar代理的方式, 转发请求, 同时实现负载限流等
	var data = make(typex.M)
	err = conn.Invoke(
		context.Background(),
		"/hello.TestApi/Version", typex.M{"input": "error"}, &data,
		//"/hello.TestApi/Version", types.M{"input": "hello"}, &data,
		grpc.ForceCodec(encoding.GetCodec("json")),
	)
	fmt.Println(err)
	fmt.Println(data)
	//select {}
}
