package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/libp2p/go-yamux/v4"
	_ "github.com/libp2p/go-yamux/v4"
	"github.com/pubgo/funk/assert"
	logger "github.com/pubgo/funk/log"
	"google.golang.org/grpc"
)

func main2() {
	listener := assert.Must1(net.Listen("tcp", ":8888"))

	// Accept a TCP connection
	conn := assert.Must1(listener.Accept())

	// Setup server side of yamux
	session := assert.Must1(yamux.Client(conn, nil, nil))

	// 验证密码，获取服务信息，版本信息
	// 验证通过之后才能真正的建立连接
	sss := assert.Must1(session.AcceptStream())
	dd := make([]byte, 1024)
	n := assert.Must1(sss.Read(dd))
	fmt.Println(string(dd[:n]))

	logger.Info().Msg("session")

	// Accept a stream

	cli := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				stream := assert.Must1(session.OpenStream(context.Background()))
				return stream, nil
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	req := assert.Must1(http.NewRequest("GET", "http://localhost:8080/hello", nil))
	rsp := assert.Must1(cli.Do(req))
	fmt.Println(string(assert.Must1(io.ReadAll(rsp.Body))))

	connCli := assert.Must1(grpc.Dial("test:8080", grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return session.Open(ctx)
	})))
	_ = connCli
}
