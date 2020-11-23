package ctl_entry

import (
	"fmt"

	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func GetEntry() golug_entry.Entry {
	ent := golug.NewCtlEntry("ctl")
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("entry http test"))

	ent.Main(func() {
		fmt.Println("ok")
	})

	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: "hello",
		OnInit: func(ent golug_entry.Entry) {
			fmt.Println("hello plugin")
		},
	}))

	return ent
}
