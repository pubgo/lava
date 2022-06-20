package bbolt

import (
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Provider(func(c config.Config, log *logging.Logger) map[string]*Client {
		return config.MakeClient(c, Name, func(key string, cfg *Cfg) *Client {
			return New(cfg.Create(), log.Named(Name))
		})
	})
}
