package runtime

import (
	"strconv"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/flagutil"
	"github.com/pubgo/lava/pkg/typex"
)

func init() {
	// 运行环境
	flags.Register(&cli.GenericFlag{
		Name:    "mode",
		Aliases: typex.StrOf("m"),
		Usage:   "running mode(dev|test|stag|prod|release)",
		EnvVars: env.KeyOf("lava_mode", "app_mode"),
		Value: flagutil.Generic{
			Value: Mode.String(),
			Destination: func(val string) error {
				var i, err = strconv.Atoi(val)
				xerror.Panic(err)

				Mode = RunMode(i)
				xerror.Assert(Mode.String() == "", "unknown mode, mode=%s", val)
				return nil
			},
		},
	})
}
