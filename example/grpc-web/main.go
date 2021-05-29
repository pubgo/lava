package main

import (
	"github.com/pubgo/lug/builder/grpc-web"
	"github.com/pubgo/lug/example/grpc_entry/handler"
	"github.com/pubgo/lug/example/proto/hello"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"net/url"

	"context"
	"encoding/json"
	"fmt"
	_ "github.com/gin-gonic/gin/binding"
	"net/http"
	"time"
	_ "unsafe"
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

type codecUri struct{}

func (c *codecUri) Marshal(v interface{}) ([]byte, error) {
	fmt.Printf("%#v\n", v)
	return json.Marshal(v)
}

func (c *codecUri) Unmarshal(data []byte, v interface{}) error {
	fmt.Println(string(data))

	var u, err = url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	return mapFormByTag(v, u, "json")
}

func (c *codecUri) Name() string {
	return "uri"
}

func init() {
	encoding.RegisterCodec(&codec{})
	encoding.RegisterCodec(&codecUri{})
}

func main() {
	grpcServer := grpc.NewServer()
	hello.RegisterTestApiServer(grpcServer, handler.NewTestAPIHandler())
	hello.RegisterTransportServer(grpcServer, &trans{})
	fmt.Println(grpcServer.GetServiceInfo())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println(r.RequestURI)
		//fmt.Println(r.Header)
		fmt.Println(r.URL.Path)
		fmt.Println(r.URL.Path)

		//q.Q(r)

		//uri, err := url.ParseQuery(r.URL.RawQuery)
		//xerror.Panic(err)
		//
		//var mm = new(interface{})
		//xerror.Panic(mapFormByTag(mm, uri, "json"))
		//q.Q(mm)

		grpcWeb.Middleware(grpcServer, w, r)

		return
	})

	http.ListenAndServe("127.0.0.1:8900", nil)
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

//go:linkname mapFormByTag github.com/gin-gonic/gin/binding.mapFormByTag
func mapFormByTag(ptr interface{}, form map[string][]string, tag string) error
