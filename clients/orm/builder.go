package orm

import (
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func New(cfg Cfg, log *logging.Logger) *Client {
	var builder = DefaultCfg()
	builder.log = log.Named(Name)

	assert.Must(merge.Struct(builder, cfg))
	assert.Must(builder.Valid())
	assert.Must(builder.Build())
	return &Client{DB: builder.Get()}
}
