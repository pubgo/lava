package orm

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/merge"
)

func init() {
	dix.Register(func(c config.Config, log *logging.Logger) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = config.Decode[*Cfg](c, Name)
		for name := range cfgMap {
			var cfg = DefaultCfg()
			xerror.Panic(merge.Struct(cfg, cfgMap[name]))
			xerror.Panic(cfg.Valid())
			clients[name] = &Client{DB: cfg.Create(log)}
		}
		return clients
	})
}
