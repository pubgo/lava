package ctl_entry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pubgo/lava"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/plugin"
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

	ent.Plugin(&plugin.Base{
		Name: "hello",
		OnInit: func(ent plugin.Entry) {
			fmt.Println("hello plugin")
		},
	})

	ent.Register(&Service{})

	return ent
}
