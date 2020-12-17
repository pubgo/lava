package ctl_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xlog"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewCtlEntry("ctl", nil)
	ent.Version("v0.0.1")
	ent.Description("entry http test")

	ent.Register(func() {
		xlog.Info("ctl ok")
	})

	golug.RegisterPlugin(&golug_plugin.Base{
		Name: "hello",
		OnInit: func(ent golug_entry.Entry) {
			fmt.Println("hello plugin")
		},
	})

	return ent
}
