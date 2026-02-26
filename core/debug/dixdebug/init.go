package dixdebug

import (
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/dix/v2/dixhttp"

	"github.com/pubgo/lava/v2/core/debug"
)

func Init(d *dix.Dix) {
	ss := dixhttp.NewServerWithOptions(d, dixhttp.WithBasePath("/debug/dix"))
	debug.Get("/dix", ss.HandleIndex)
	debug.Get("/dix/api/dependencies", ss.HandleDependencies)
	debug.Get("/dix/api/stats", ss.HandleStats)
	debug.Get("/dix/api/packages", ss.HandlePackages)
	debug.Get("/dix/api/package", ss.HandlePackageDetails)
	debug.Get("/dix/api/type", ss.HandleTypeDetails)
	debug.Get("/dix/api/group-rules", ss.HandleGroupRules)
}
