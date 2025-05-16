package tunnelserver

import (
	"github.com/goccy/go-json"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/rpc/jsonrpc"
	"time"

	yamux "github.com/libp2p/go-yamux/v5"
	_ "github.com/xtaci/smux"
	_ "golang.ngrok.com/muxado/v2"
)

func client() {
	// Get a TCP connection
	conn, err := net.Dial(...)
	if err != nil {
		panic(err)
	}

	// Setup client side of yamux
	session, err := yamux.Client(conn, nil, nil)
	if err != nil {
		panic(err)
	}

	// Open a new stream
	stream, err := session.Open(nil)
	if err != nil {
		panic(err)
	}
	stream.Close()

	// Stream implements net.Conn
	stream.Write([]byte("ping"))
}

func server() {
	http.DefaultServeMux.HandleFunc("", func(writer http.ResponseWriter, request *http.Request) {
		httputil.NewSingleHostReverseProxy(nil).ServeHTTP(writer, request)
	})
	srv := &http.Server{}

	listener, _ := net.Listen("", "")

	for {
		// Accept a TCP connection
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			// Setup server side of yamux
			session, err := yamux.Server(conn, nil, nil)
			if err != nil {
				panic(err)
			}

			// Accept a stream
			stream, err := session.Accept()
			if err != nil {
				panic(err)
			}

			sss := jsonrpc.NewServerCodec(stream)
			for {
				sss.ReadRequestHeader()
				sss.WriteResponse(nil, nil)
			}

			ccc := jsonrpc.NewClientCodec(stream)
			ccc.WriteRequest()
			ccc.ReadResponseHeader()

			json.NewDecoder(stream).Decode()
			json.NewEncoder(stream).Encode()
			stream.Close()

			// Listen for a message
			buf := make([]byte, 4)
			stream.Read(buf)
			stream.Write(buf)
		}()
	}

	//session.Open(nil)
	//
	//http.Client{
	//	Transport: http.DefaultTransport,
	//}
	//
	//_ = httpproxy.Config{}
	//proxy.Dial()
	//proxy.Dial()
	//session.Open(nil)
}

func init() {
	_ = yamux.Config{}
	// 关键参数调整
	yamuxConfig := yamux.DefaultConfig()
	yamuxConfig.MaxStreamWindowSize = 1024 * 1024 * 4 // 单个流窗口 4MB
	yamuxConfig.AcceptBacklog = 128                   // 流接收队列长度
	yamuxConfig.EnableKeepAlive = true                // 开启保活
	yamuxConfig.KeepAliveInterval = 15 * time.Second  // 保活间隔
	yamuxConfig.MaxMessageSize = 16 * 1024 * 1024     // 最大消息大小
	yamuxConfig.LogOutput = io.Discard                // 日志输出
	yamux.Server()
}
