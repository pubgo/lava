package dixdebug

import (
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixhttp"

	"github.com/pubgo/lava/v2/core/debug"
)

func Init(d *dix.Dix) {
	ss := dixhttp.NewServerWithOptions(d, dixhttp.WithBasePath("/debug"))
	debug.Get("/dix", ss.HandleIndex)
	debug.Get("/api/dependencies", ss.HandleDependencies)
	debug.Get("/api/stats", ss.HandleStats)
	debug.Get("/api/packages", ss.HandlePackages)
	debug.Get("/api/package/", ss.HandlePackageDetails)
	debug.Get("/api/type/", ss.HandleTypeDetails)
}
