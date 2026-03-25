package dixdebug

import (
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixhttp"

	"github.com/pubgo/lava/v2/core/debug"
)

func Init(d *dix.Dix) {
	debug.Group("/dix", dixhttp.NewServerWithOptions(d, dixhttp.WithBasePath("/debug/dix")))
}
