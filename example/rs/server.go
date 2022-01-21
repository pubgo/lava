package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/core"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugins/syncx"
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

	ln := xerror.PanicErr(netutil.Listen(":7878")).(net.Listener)
	mux := cmux.New(ln)
	mux.SetReadTimeout(time.Second * 2)
	mux.HandleError(func(err error) bool {
		logutil.LogOrErr(logging.L(), "mux error", func() error { return err })
		return false
	})

	var sln = mux.Match(func(r io.Reader) bool {
		br := bufio.NewReader(&io.LimitedReader{R: r, N: 4096})
		l, part, err := br.ReadLine()
		if err != nil || part {
			logutil.LogOrErr(logging.L(), "ReadLine", func() error { return err })
			return false
		}

		logging.L().Debug(string(l))

		// 用于websocket匹配
		if cmux.HTTP1()(bytes.NewBuffer(l)) {
			return true
		}

		// 用于rsocket匹配
		var frame = transport.NewLengthBasedFrameDecoder(bytes.NewBuffer(l))
		data, err := frame.Read()
		if err != nil {
			logutil.LogOrErr(logging.L(), "frame.Read", func() error { return err })
			return false
		}

		var header = core.ParseFrameHeader(data)
		if header.Type().String() == "UNKNOWN" {
			return false
		}

		logging.L().Debug(header.String())

		return true
	})

	syncx.GoDelay(func() {
		xerror.Panic(mux.Serve())
	})

	_, err1 := quic.ListenAddr(":7878", &tls.Config{InsecureSkipVerify: true}, nil)
	xerror.Panic(err1)

	err := rsocket.Receive().
		OnStart(func() {
			logging.L().Info("Server Start")
		}).
		Acceptor(func(ctx context.Context, setup payload.SetupPayload, clientSocket rsocket.CloseableRSocket) (rsocket.RSocket, error) {
			logging.L().Info(
				"setup",
				zap.String("data", string(setup.Data())),
				zap.String("version", setup.Version().String()),
			)

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
		}).

		// specify transport
		//Transport(rsocket.TCPServer().SetAddr(":7878").Build()).
		Transport(func(ctx context.Context) (transport.ServerTransport, error) {
			return transport.NewWebsocketServerTransport(func(ctx context.Context) (net.Listener, error) {
				return sln, nil
			}, "/hello", nil), nil

			//return transport.NewTCPServerTransport(func(ctx context.Context) (net.Listener, error) {
			//	return sln, nil
			//}), nil
		}).
		// serve will block execution unless an error occurred
		Serve(context.Background())

	panic(err)
}

// wordCount function
func wordCount(value string) int {
	words := strings.Fields(value)
	return len(words)
}
