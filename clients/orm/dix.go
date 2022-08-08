package orm

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func init() {
	defer recovery.Exit()

	dix.Provider(func(c config.Config, log *logging.Logger) map[string]*Client {
		return config.MakeClient(c, Name, func(key string, cfg *Cfg) *Client {
			var builder = DefaultCfg()
			builder.log = log.Named(Name)

			assert.Must(merge.Struct(builder, cfg))
			assert.Must(builder.Valid())
			assert.Must(builder.Build())
			return &Client{DB: builder.Get()}
		})
	})
}
