package util

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"

	"github.com/pubgo/lava/cmd/lava-broker/rs_manager"
	"github.com/pubgo/lava/cmd/lava-broker/rsocket/sockets"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logutil"
)

type serverCfg struct {

	//wait serverCfg run success
	wg *sync.WaitGroup

	//given serverName to service discovery to find
	serverName string

	//tcp socket address
	tcpAddress string

	//websocket address
	wssAddress string

	//requestChannel buffSize setting
	buffSize int

	//rsocket serverBuilder
	serverBuilder rsocket.ServerBuilder

	//rsocket serverStarter
	serverStart rsocket.ToServerStarter

	transport transport.ServerTransport
}

func (r *serverCfg) Address() string {
	return "[tcp: " + r.tcpAddress + "] [wss: " + r.wssAddress + "]"
}

func (r *serverCfg) String() string {
	return "rsocket"
}

func NewServer(tcpAddress, wssAddress, serverName string, buffSize int) *serverCfg {
	return &serverCfg{
		serverName: serverName,
		tcpAddress: tcpAddress,
		wssAddress: wssAddress,
		buffSize:   buffSize,
	}
}

func (r *serverCfg) Build(options ...func(opts *sockets.Handler)) error {
	var srv = rsocket.Receive()
	srv.Resume()
	// setting scheduler goroutine on numCPU*2 to better working
	srv.Scheduler(scheduler.NewElastic(runtime.NumCPU()<<8), scheduler.NewElastic(runtime.NumCPU()<<8))

	srv.OnStart(func() {
		logging.L().Info("serverCfg start")
	})

	return srv.Acceptor(func(ctx context.Context, setup payload.SetupPayload, socket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
		// 客户端地址
		var clientIp, ok = rsocket.GetAddr(socket)
		if !ok || clientIp == "" {
			return nil, fmt.Errorf("get socket address info failed")
		}

		ctx, cancel := context.WithCancel(ctx)

		// 客户端关闭
		socket.OnClose(func(err error) {
			cancel()
			logutil.LogOrErr(logging.L(), "socket close", func() error { return err })
		})

		fmt.Println("clientIp=>", clientIp)

		// 服务名字
		// 版本
		// 绑定客户端的信息

		// 用于服务端和客户端版本协调
		_ = setup.Version()

		// 默认metadata的编码方式
		_ = setup.MetadataMimeType()

		// 默认data的编码方式
		_ = setup.DataMimeType()

		setup.TimeBetweenKeepalive()
		_, _ = setup.Metadata()

		var srv rs_manager.Service
		switch srv.Kind {
		case "server":
			// TODO 返回server handler
			// 获取服务信息, 获取注册信息, 获取节点信息
		case "client":
			// TODO 返回client handler
			// 要能够获取请求的服务信息，然后，转发信息到注册到服务端
			// 服务端，进行负载均衡的访问
		}

		// 客户端启动注册的信息
		_ = setup.Data()
		// TODO 检查client和server

	}).Transport(func(ctx context.Context) (transport.ServerTransport, error) { return r.transport, nil }).Serve(nil)
}
