package grpcall

import (
	"context"
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var args = func(cmd *cobra.Command) *cobra.Command {
		return cmd
	}

	return args(&cobra.Command{
		Use: "grpcall",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args)
			cc, err := BlockingDial(context.Background(), args[0])
			refClient := grpcreflect.NewClient(context.Background(), reflectpb.NewServerReflectionClient(cc))
			descSource := DescriptorSourceFromServer(context.Background(), refClient)
			fmt.Println(descSource.ListServices())
			fmt.Println(descSource.FindSymbol(args[1]))
			dsc, _ := descSource.FindSymbol(args[1])
			sd, _ := dsc.(*desc.ServiceDescriptor)
			mtd := sd.FindMethodByName(args[2])
			fmt.Println(mtd)

			cc1, _ := New()

			ff, err := ParseFormatterByDesc(descSource, true)
			xerror.Panic(err)
			SetDefaultEventHandler(descSource, ff)
			cc1.eventHandler = defaultInEventHooker

			var handler = &DefaultEventHandler1{}
			cc1.invokeCtl = newInvokeHandler(DefaultEventHandler, handler)

			res, err := cc1.Call(args[0], args[1], args[2], `{"input":"hello"}`)
			time.Sleep(time.Millisecond * 10)
			fmt.Printf("%#v \n", res)
			fmt.Printf("%#v \n", err)

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func() {
				defer wg.Done()
				for {
					select {
					case res.SendChan <- []byte(`{"input":"hello"}`):
						fmt.Println("send")
						time.Sleep(time.Second)

					case <-res.DoneChan:
						return
					}
				}
			}()

			go func() {
				defer wg.Done()

				for {
					select {
					case msg, ok := <-res.ResultChan:
						fmt.Println("recv data:", msg, ok)
					case err := <-res.DoneChan:
						fmt.Println("done chan: ", err)
						return
					}
				}
			}()

			wg.Wait()

		},
	})
}

type DefaultEventHandler1 struct {
	sendChan chan []byte
}

func (h *DefaultEventHandler1) OnReceiveData(md metadata.MD, resp string, respErr error) {
}

func (h *DefaultEventHandler1) OnReceiveTrailers(stat *status.Status, md metadata.MD) {
}

//-d {"input":"hello"}    localhost:8080 hello.TestApi/Version
