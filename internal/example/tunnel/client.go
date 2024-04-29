package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-yamux/v4"
	"github.com/pubgo/funk/assert"
	"net"
	"net/http"
)

func main1() {
	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})

	// Get a TCP connection
	conn := assert.Must1(net.Dial("tcp", ":8888"))

	// Setup client side of yamux
	session := assert.Must1(yamux.Server(conn, nil, nil))

	ssss := assert.Must1(session.OpenStream(context.Background()))
	ssss.Write([]byte("service name, secret, 版本信息等"))

	// Open a new stream
	//stream := assert.Must1(session.AcceptStream())
	//defer stream.Close()
	//fmt.Println(stream.StreamID())

	server := &http.Server{}
	fmt.Println(server.Serve(session))
}
