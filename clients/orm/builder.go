package orm

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/result"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
)

func New(cfg *Cfg, log *logging.Logger) *Client {
	assert.If(cfg == nil, "config is nil")

	var builder = DefaultCfg()
	builder.log = log.Named(Name)
	builder = merge.Struct(builder, cfg).Unwrap(func(err result.Error) result.Error {
		return err.WrapF("cfg=%#v", cfg)
	})
	assert.Must(builder.Build())
	return &Client{DB: builder.Get()}
}
