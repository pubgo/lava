package orm

import (
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
)

func New(cfg *Cfg, log *logging.Logger) *Client {
	assert.If(cfg == nil, "config is nil")

	var builder = DefaultCfg()
	builder.log = log.Named(Name)

	assert.Must(merge.Struct(builder, cfg))
	assert.Must(builder.Valid())
	assert.Must(builder.Build())
	return &Client{DB: builder.Get()}
}
