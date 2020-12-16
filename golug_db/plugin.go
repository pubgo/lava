package golug_db

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Panic(golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent golug_entry.Entry) {
			ent.Decode(Name, &cfg)

			for k, v := range cfg {
				_cfg := GetDefaultCfg()
				golug_util.Mergo(&_cfg, v)

				initClient(k, _cfg)
				cfg[k] = _cfg
			}
		},
	}))
}
