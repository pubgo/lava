package config

import (
	"sort"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	defer recovery.Exit()

	vars.Register("config_data", func() interface{} {
		var conf = dix.Inject(new(struct{ C Config }))
		return conf.C.All()
	})

	vars.Register("config_keys", func() interface{} {
		var conf = dix.Inject(new(struct{ C Config }))
		var keys = conf.C.AllKeys()
		sort.Strings(keys)
		return keys
	})

	vars.Register("config", func() interface{} {
		return typex.M{
			"cfg_type": FileType,
			"cfg_name": FileName,
			"home":     CfgDir,
			"cfg_path": CfgPath,
		}
	})
}
