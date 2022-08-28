package config

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register("config", func() interface{} {
		return typex.M{
			"cfg_type": FileType,
			"cfg_name": FileName,
			"home":     CfgDir,
			"cfg_path": CfgPath,
		}
	})
}
