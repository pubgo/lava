package ctl_entry

import (
	"fmt"
	"time"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xlog"
	"github.com/spf13/cobra"
)

var name = "test-ctl"
var logs = xlog.GetLogger(name)

func GetEntry() entry.Entry {
	ent := lug.NewCtl(name)
	ent.Description("entry ctl test")
	ent.Commands(&cobra.Command{
		Use: "sub",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("sub cmd")
		},
	})

	ent.Register(consts.Default, func(ctx fx.Ctx) {
		logs.Info("ctl ok")
	})

	ent.RegisterLoop("hello", func(ctx fx.Ctx) {
		logs.Info("ctl hello")
		time.Sleep(time.Second)
	})

	ent.Plugin(&plugin.Base{
		Name: "hello",
		OnInit: func(ent entry.Entry) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
