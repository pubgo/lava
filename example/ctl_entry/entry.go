package ctl_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/entry"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xlog"
)

var name = "test-ctl"

func GetEntry() entry.Entry {
	ent := golug.NewCtl(name)
	ent.Version("v0.0.1")
	ent.Description("entry ctl test")

	ent.Register(func() {
		xlog.Info("ctl ok")
	})

	golug.Plugin(&plugin.Base{
		Name: "hello",
		OnInit: func(fn interface{}) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
