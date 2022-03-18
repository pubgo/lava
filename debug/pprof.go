package debug

import (
	"net/http/pprof"

	"github.com/felixge/fgprof"
	adaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

func init() {
	Get("/gprof", adaptor.HTTPHandler(fgprof.Handler()))
	Route("/pprof", func(r fiber.Router) {
		r.Get("/", adaptor.HTTPHandlerFunc(pprof.Index))
		r.Get("/cmdline", adaptor.HTTPHandlerFunc(pprof.Cmdline))
		r.Get("/profile", adaptor.HTTPHandlerFunc(pprof.Profile))
		r.Get("/symbol", adaptor.HTTPHandlerFunc(pprof.Symbol))
		r.Get("/trace", adaptor.HTTPHandlerFunc(pprof.Trace))
		r.Get("/allocs", adaptor.HTTPHandler(pprof.Handler("allocs")))
		r.Get("/goroutine", adaptor.HTTPHandler(pprof.Handler("goroutine")))
		r.Get("/heap", adaptor.HTTPHandler(pprof.Handler("heap")))
		r.Get("/mutex", adaptor.HTTPHandler(pprof.Handler("mutex")))
		r.Get("/threadcreate", adaptor.HTTPHandler(pprof.Handler("threadcreate")))
	})
}
