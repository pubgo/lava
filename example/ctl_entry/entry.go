package ctl_entry

import (
	"fmt"

	"github.com/pubgo/lug"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xlog"
)

var name = "test-ctl"

func GetEntry() entry.Abc {
	ent := lug.NewCtl(name)
	ent.Version("v0.0.1")
	ent.Description("entry ctl test")

	ent.Register(func() {
		xlog.Info("ctl ok")
	})

	lug.Plugin(&plugin.Base{
		Name: "hello",
		OnInit: func(fn interface{}) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
