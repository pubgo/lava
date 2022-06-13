package orm

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/merge"
)

func init() {
	dix.Register(func(c config.Config) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = config.Decode[*Cfg](c, Name)
		for name, cfg := range cfgMap {
			xerror.Panic(merge.Struct(&cfg, DefaultCfg()))
			xerror.Panic(cfg.Valid())
			clients[name] = &Client{DB: cfg.Create()}
		}
		return clients
	})
}
