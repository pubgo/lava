package configcmd

import (
	"context"
	"fmt"

	"github.com/pubgo/redant"
	yaml "gopkg.in/yaml.v3"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/recovery"
)

func New[Cfg any]() *redant.Command {
	return &redant.Command{
		Use:   "config",
		Short: "config management",
		Children: []*redant.Command{
			{
				Use:   "show",
				Short: "show config data",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					defer recovery.Exit()
					fmt.Println("config path:\n", config.GetConfigPath())
					fmt.Println("config raw data:\n", string(assert.Must1(yaml.Marshal(config.Load[Cfg]().T))))
					return nil
				},
			},
		},
	}
}
