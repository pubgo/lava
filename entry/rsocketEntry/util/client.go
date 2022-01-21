package util

import (
	"runtime"
	"time"

	"github.com/jjeffcaii/reactor-go/scheduler"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"

	"github.com/pubgo/lava/endpoint"
	"github.com/pubgo/lava/logger"
)

//this is rsocket clientCfg
type clientCfg struct {

	//rsocket client
	client rsocket.Client

	//rsocket connect timeout
	connectTimeout time.Duration

	//rsocket keepalive interval
	keepaliveInterval time.Duration

	//rsocket keepalive life time
	keepaliveLifetime time.Duration
}

func NewClient(connTimeout, interval, tll time.Duration) *clientCfg {
	return &clientCfg{
		connectTimeout:    connTimeout,
		keepaliveInterval: interval,
		keepaliveLifetime: tll,
	}
}

func (cli *clientCfg) Build(e *endpoint.Endpoint, ch chan string) (rsocket.ClientBuilder, error) {
	var client = rsocket.Connect()
	//client.MetadataMimeType(extension.ApplicationProtobuf.String()).
	//client.DataMimeType(extension.ApplicationProtobuf.String()).

	//set scheduler to best
	client.Scheduler(scheduler.NewElastic(runtime.NumCPU()<<8), scheduler.NewElastic(runtime.NumCPU()<<8))
	client.KeepAlive(cli.keepaliveInterval, cli.keepaliveLifetime, 1)
	client.ConnectTimeout(cli.connectTimeout)

	// 连接服务器后发送的消息
	// 服务注册消息
	client.SetupPayload(payload.New(nil, nil))

	//handler when connect success
	client.OnConnect(func(client rsocket.Client, err error) {
		logger.S().Debugf("connected at: %s", e.Address)
	})

	//when net occur some error,it's will be callback the error server ip address
	client.OnClose(func(err error) {
		if err != nil {
			logger.S().Errorf("server [%s %s] is closed |err=%v", e.Name, e.Address, err)
		} else {
			logger.S().Debugf("server [%s %s] is closed", e.Name, e.Address)
		}

		ch <- e.Address
	})

	return client, nil
}
