package config

import (
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	defer recovery.Exit()

	vars.Register("config", func() interface{} {
		return typex.M{
			"cfg_type": FileType,
			"cfg_name": FileName,
			"home":     CfgDir,
			"cfg_path": CfgPath,
		}
	})
}
