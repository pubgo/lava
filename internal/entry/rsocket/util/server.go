package util

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"runtime"
	"sync"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
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

func (r *serverCfg) Build(rs rsocket.RSocket) rsocket.ServerBuilder {
	var srv = rsocket.Receive()
	srv.OnStart(func() {
		logging.L().Info("serverCfg start")
	})

	srv.Resume()

	// setting scheduler goroutine on numCPU*2 to better working
	srv.Scheduler(scheduler.NewElastic(runtime.NumCPU()<<8), scheduler.NewElastic(runtime.NumCPU()<<8))

	srv.Acceptor(func(ctx context.Context, setup payload.SetupPayload, socket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
		ctx, cancel := context.WithCancel(ctx)

		// 客户端关闭
		socket.OnClose(func(err error) {
			cancel()
			logutil.LogOrErr(logging.L(), "client close", func() error { return err })
		})

		// 客户端地址
		var remoteIp, ok = rsocket.GetAddr(socket)
		if !ok {
			return nil, fmt.Errorf("get address info failed")
		}
		fmt.Println("remoteIp=>", remoteIp)

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

		// 客户端启动注册的信息
		_ = setup.Data()

		return rs, nil
	})

	return srv
}
