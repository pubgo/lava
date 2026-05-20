package envcmd

import (
	"context"
	"fmt"

	"github.com/pubgo/funk/v2/config"
	"github.com/pubgo/funk/v2/env"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "envs",
		Short: "show all envs",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			defer recovery.Exit()

			env.Reload()

			fmt.Println("config path:", config.GetConfigPath())
			envs := config.LoadEnvMap(config.GetConfigPath())
			for name, cfg := range envs {
				envData := env.Get(name)
				if envData != "" {
					cfg.Default = envData
				}
			}

			pretty.Println(envs)
			return nil
		},
	}
}
