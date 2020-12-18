package golug_badger

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/internal/pkg/golug_util"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Panic(golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent golug_entry.Entry) {
			golug_config.Decode(Name, &cfg)

			for k, v := range cfg {
				_cfg := GetDefaultCfg()
				golug_util.Mergo(&_cfg, v)

				initClient(k, _cfg)
				cfg[k] = _cfg
			}
		},
	}))
}
