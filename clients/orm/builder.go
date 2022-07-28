package orm

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func init() {
	defer recovery.Exit()

	dix.Provider(func(c config.Config, log *logging.Logger) map[string]*Client {
		return config.MakeClient(c, Name, func(key string, cfg *Cfg) *Client {
			var dc = DefaultCfg()
			xerror.Panic(merge.Struct(dc, cfg))
			xerror.Panic(cfg.Valid())
			return &Client{DB: cfg.Create(log)}
		})
	})
}
