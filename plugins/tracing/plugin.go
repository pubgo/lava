package tracing

import (
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnVars: func(v types.Vars) {
			v.Do(Name+"_cfg", func() interface{} { return cfg })
			v.Do(Name+"_factory", func() interface{} {
				var data = make(map[string]string)
				factories.Range(func(key, value interface{}) bool {
					data[key.(string)] = stack.Func(value)
					return true
				})
				return data
			})
		},
	})
}
