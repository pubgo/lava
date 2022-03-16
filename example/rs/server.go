package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pubgo/lava/pkg/syncx"
	"net"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
)

func main() {
	server()
	time.Sleep(time.Second)
}

func server() {
	// create a srv that will be called when the server receives the RequestChannel frame (FrameTypeRequestChannel - 0x07)
	requestChannelHandler := rsocket.RequestChannel(func(requests flux.Flux) flux.Flux {
		return flux.Create(func(ctx context.Context, s flux.Sink) {
			requests.DoOnNext(func(elem payload.Payload) error {
				// for each payload in a flux stream respond with a word count
				s.Next(payload.NewString(fmt.Sprintf("%d", wordCount(elem.DataUTF8())), ""))
				return nil
			}).DoOnComplete(func() {
				// signal completion of the response stream
				s.Complete()
			}).Subscribe(context.Background())
		})
	})

	var cfg = netutil.DefaultCfg()
	cfg.Port = 7878
	cfg.HandleError = func(err error) bool {
		logutil.LogOrErr(logging.L(), "mux error", func() error { return err })
		return true
	}

	var lnFn = cfg.Rsocket()
	var wsFn = cfg.Websocket()

	syncx.GoDelay(func() {
		xerror.Panic(cfg.Serve())
	})

	_, err1 := quic.ListenAddr(":7878", &tls.Config{InsecureSkipVerify: true}, nil)
	xerror.Panic(err1)

	var ts = rsocket.Receive().
		OnStart(func() {
			logging.L().Info("Server Start")
		}).
		Acceptor(func(ctx context.Context, setup payload.SetupPayload, clientSocket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			logging.L().Info(
				"setup",
				zap.String("data", string(setup.Data())),
				zap.String("version", setup.Version().String()),
			)

			if string(setup.Data()) != "hello123" {
				clientSocket.FireAndForget(payload.New([]byte("认证失败"), nil))
				xerror.Panic(clientSocket.Close())
				return nil, fmt.Errorf("client close")
			}

			go func() {
				// 向客户端发送一条消息
				i := 0
				for range time.Tick(time.Second) {
					i++

					if i == 2 {
						clientSocket.FireAndForget(payload.New([]byte("close"), nil))
					}
					clientSocket.FireAndForget(payload.New([]byte("check"), nil))
				}
			}()

			clientSocket.OnClose(func(err error) {
				logutil.LogOrErr(logging.L(), "server: Client Close", func() error { return err })
			})

			// 主动关闭客户端
			// clientSocket.Close()

			// register a new request channel srv
			return rsocket.NewAbstractSocket(requestChannelHandler), nil
		})

	// specify transport
	go func() {
		xerror.Panic(ts.Transport(func(ctx context.Context) (transport.ServerTransport, error) {
			return transport.NewWebsocketServerTransport(
				func(ctx context.Context) (net.Listener, error) { return wsFn(), nil },
				"/hello", nil), nil
		}).Serve(context.Background()))
	}()

	go func() {
		xerror.Panic(ts.Transport(func(ctx context.Context) (transport.ServerTransport, error) {
			return transport.NewTCPServerTransport(func(ctx context.Context) (net.Listener, error) {
				return lnFn(), nil
			}), nil
		}).Serve(context.Background()))
	}()

	select {}
}

// wordCount function
func wordCount(value string) int {
	words := strings.Fields(value)
	return len(words)
}
