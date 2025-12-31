package dixdebug

import (
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixhttp"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/pkg/httputil"
)

func Init(d *dix.Dix) {
	ss := dixhttp.NewServer(d)
	debug.Get("/dix", debug.WrapFunc(ss.HandleIndex))
	debug.Get("/api/dependencies", httputil.StripPrefix("/debug", debug.WrapFunc(ss.HandleDependencies)))
}
