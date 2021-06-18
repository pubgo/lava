package ctl_entry

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"time"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/entry/ctl"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xlog"
)

var name = "test-ctl"

func GetEntry() entry.Entry {
	ent := lug.NewCtl(name)
	ent.Description("entry ctl test")
	ent.Commands(&cobra.Command{
		Use: "sub",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sub cmd")
		},
	})

	ent.Register(func(ctx context.Context) {
		xlog.Info("ctl ok")
	})

	ent.RegisterLoop(func(ctx context.Context) {
		xlog.Info("ctl hello")
		time.Sleep(time.Second)
	}, ctl.WithName("hello"))

	ent.Plugin(&plugin.Base{
		Name: "hello",
		OnInit: func(fn interface{}) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
