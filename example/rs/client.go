package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/balancer"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

func main() {
	var _ = balancer.NewGroup(func() balancer.Balancer {
		return balancer.NewRoundRobinBalancer()
	})

	client()
	client1()
	time.Sleep(time.Second)
}

func client() {
	// Start a client connection
	client, err := rsocket.
		Connect().
		SetupPayload(payload.New([]byte("hello123"), nil)).
		OnConnect(func(c rsocket.Client, err error) {
			logutil.LogOrErr(logging.L(), "Client Connect", func() error { return err })
		}).
		OnClose(func(err error) {
			logutil.LogOrErr(logging.L(), "Client Close", func() error { return err })
		}).Acceptor(func(ctx context.Context, client rsocket.RSocket) rsocket.RSocket {
		if cr, ok := client.(rsocket.CloseableRSocket); ok {
			// TODO 客户端主动关闭服务
			// 一般用于服务更新的时候, 客户端切换到新服务中, 切换过程中, 客户端几乎无感知
			_ = cr
		}

		// 双向流, 接收从服务端主动发送的消息
		//	服务端RSocket实现
		return rsocket.NewAbstractSocket(rsocket.FireAndForget(func(request payload.Payload) {
			if string(request.Data()) == "close" {
				client.(rsocket.CloseableRSocket).Close()
				return
			}

			logging.L().Info(string(request.Data()))
		}))
	}).
		//Transport(rsocket.TCPClient().SetHostAndPort("127.0.0.1", 7878).Build()).Start(context.Background())
		Transport(rsocket.WebsocketClient().SetURL("ws://127.0.0.1:7878/hello").Build()).Start(context.Background())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// strings to count the words
	sentences := []payload.Payload{
		payload.NewString("", extension.TextPlain.String()),
		payload.NewString("qux", extension.TextPlain.String()),
		payload.NewString("The quick brown fox jumps over the lazy dog", extension.TextPlain.String()),
		payload.NewString("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", extension.TextPlain.String()),
	}

	f := flux.FromSlice(sentences)

	// create a wait group so that the function does not return until the stream completes
	wg := sync.WaitGroup{}
	wg.Add(1)

	counter := 0

	// register srv for RequestChannel
	client.RequestChannel(f).DoOnNext(func(input payload.Payload) error {
		// print word count
		fmt.Println(sentences[counter].DataUTF8(), ":", input.DataUTF8())
		counter = counter + 1
		time.Sleep(time.Second)
		return nil
	}).DoOnComplete(func() {
		// will be called on successful completion of the stream
		fmt.Println("Word counter ended.")
	}).DoOnError(func(err error) {
		// will be called if a error occurs
		fmt.Println(err)
	}).DoFinally(func(s rx.SignalType) {
		// will always be called
		wg.Done()
	}).Subscribe(context.Background())

	// wait until the stream has finished
	wg.Wait()
}

func client1() {
	// Start a client connection
	client, err := rsocket.
		Connect().
		SetupPayload(payload.New([]byte("hello123"), nil)).
		OnConnect(func(c rsocket.Client, err error) {
			logutil.LogOrErr(logging.L(), "Client Connect", func() error { return err })
		}).
		OnClose(func(err error) {
			logutil.LogOrErr(logging.L(), "Client Close", func() error { return err })
		}).Acceptor(func(ctx context.Context, client rsocket.RSocket) rsocket.RSocket {
		if cr, ok := client.(rsocket.CloseableRSocket); ok {
			// TODO 客户端主动关闭服务
			// 一般用于服务更新的时候, 客户端切换到新服务中, 切换过程中, 客户端几乎无感知
			_ = cr
		}

		// 双向流, 接收从服务端主动发送的消息
		//	服务端RSocket实现
		return rsocket.NewAbstractSocket(rsocket.FireAndForget(func(request payload.Payload) {
			if string(request.Data()) == "close" {
				client.(rsocket.CloseableRSocket).Close()
				return
			}

			logging.L().Info(string(request.Data()))
		}))
	}).
		Transport(rsocket.TCPClient().SetHostAndPort("127.0.0.1", 7878).Build()).Start(context.Background())
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// strings to count the words
	sentences := []payload.Payload{
		payload.NewString("", extension.TextPlain.String()),
		payload.NewString("qux", extension.TextPlain.String()),
		payload.NewString("The quick brown fox jumps over the lazy dog", extension.TextPlain.String()),
		payload.NewString("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", extension.TextPlain.String()),
	}

	f := flux.FromSlice(sentences)

	// create a wait group so that the function does not return until the stream completes
	wg := sync.WaitGroup{}
	wg.Add(1)

	counter := 0

	// register srv for RequestChannel
	client.RequestChannel(f).DoOnNext(func(input payload.Payload) error {
		// print word count
		fmt.Println(sentences[counter].DataUTF8(), ":", input.DataUTF8())
		counter = counter + 1
		time.Sleep(time.Second)
		return nil
	}).DoOnComplete(func() {
		// will be called on successful completion of the stream
		fmt.Println("Word counter ended.")
	}).DoOnError(func(err error) {
		// will be called if a error occurs
		fmt.Println(err)
	}).DoFinally(func(s rx.SignalType) {
		// will always be called
		wg.Done()
	}).Subscribe(context.Background())

	// wait until the stream has finished
	wg.Wait()
}
