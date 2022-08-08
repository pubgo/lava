package bbolt

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
)

func init() {
	defer recovery.Exit()
	dix.Provider(func(c config.Config, log *logging.Logger) map[string]*Client {
		return config.MakeClient(c, Name, func(key string, cfg *Cfg) *Client {
			assert.Must(cfg.Build())
			return New(cfg.Get(), log.Named(Name))
		})
	})
}
