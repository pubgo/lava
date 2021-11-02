package tracing

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func() {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
	})
}
