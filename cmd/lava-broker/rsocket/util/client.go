package util

import (
	"runtime"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

// ClientCfg client config
type ClientCfg struct {
	// ConnectTimeout rsocket connect timeout
	ConnectTimeout time.Duration

	// KeepaliveInterval rsocket keepalive interval
	KeepaliveInterval time.Duration

	// KeepaliveLifetime rsocket keepalive life time
	KeepaliveLifetime time.Duration

	Setup payload.Payload

	Socket rsocket.ClientSocketAcceptor

	OnClose   func(err error)
	OnConnect func(rsocket.Client, error)
}

func (t *ClientCfg) Build() rsocket.ClientBuilder {
	var client = rsocket.Connect()

	// 默认metadata数据类型
	client.MetadataMimeType(extension.ApplicationProtobuf.String())
	// 默认data数据类型
	client.DataMimeType(extension.ApplicationProtobuf.String())

	// set scheduler to best
	client.Scheduler(scheduler.NewElastic(runtime.NumCPU()<<8), scheduler.NewElastic(runtime.NumCPU()<<8))
	client.KeepAlive(t.KeepaliveInterval, t.KeepaliveLifetime, 1)
	client.ConnectTimeout(t.ConnectTimeout)

	// 连接服务器后发送的消息
	// 服务注册消息到
	// 认证信息等
	client.SetupPayload(t.Setup)

	client.OnConnect(func(client rsocket.Client, err error) {

	})

	// handler when connect success
	client.OnConnect(t.OnConnect)

	// when net occur some error,it's will be callback the error serverCfg ip address
	client.OnClose(t.OnClose)

	client.Acceptor(t.Socket)

	return client
}
