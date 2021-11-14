package cli_entry

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/entry/cliEntry"
	"github.com/pubgo/lava/plugin"
)

var name = "test-cliEntry"

func GetEntry() entry.Entry {
	ent := cliEntry.New(name)
	ent.Description("entry cliEntry test")
	ent.Commands(&cli.Command{
		Name: "sub",
		Action: func(context *cli.Context) error {
			fmt.Println("sub cmd")
			return nil
		},
	})

	plugin.Register(&plugin.Base{
		Name: "hello",
		OnInit: func(p plugin.Process) {
			fmt.Println("hello plugin")
		},
	})

	ent.Register(&Service{})

	return ent
}
