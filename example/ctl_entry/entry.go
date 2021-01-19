package ctl_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xlog"
)

var name = "test-ctl"

func GetEntry() golug_entry.Entry {
	ent := golug.NewCtlEntry(name)
	ent.Version("v0.0.1")
	ent.Description("entry ctl test")

	ent.Register(func() {
		xlog.Info("ctl ok")
	})

	golug.RegisterPlugin(&golug_plugin.Base{
		Name: "hello",
		OnInit: func(fn interface{}) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
