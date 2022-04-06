package debug

import (
	"errors"
	"net/http/pprof"

	"github.com/felixge/fgprof"
	adaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

func init() {
	Get("/gprof", adaptor.HTTPHandler(fgprof.Handler()))
	Route("/pprof", func(r fiber.Router) {
		r.Get("/", adaptor.HTTPHandlerFunc(pprof.Index))
		r.Get("/:name", func(ctx *fiber.Ctx) error {
			var name = ctx.Params("name")
			switch name {
			case "cmdline":
				return adaptor.HTTPHandlerFunc(pprof.Cmdline)(ctx)
			case "profile":
				return adaptor.HTTPHandlerFunc(pprof.Profile)(ctx)
			case "symbol":
				return adaptor.HTTPHandlerFunc(pprof.Symbol)(ctx)
			case "trace":
				return adaptor.HTTPHandlerFunc(pprof.Trace)(ctx)
			case "allocs":
				return adaptor.HTTPHandler(pprof.Handler("allocs"))(ctx)
			case "goroutine":
				return adaptor.HTTPHandler(pprof.Handler("goroutine"))(ctx)
			case "heap":
				return adaptor.HTTPHandler(pprof.Handler("heap"))(ctx)
			case "mutex":
				return adaptor.HTTPHandler(pprof.Handler("mutex"))(ctx)
			case "threadcreate":
				return adaptor.HTTPHandler(pprof.Handler("threadcreate"))(ctx)
			}

			return errors.New("name not found")
		})
	})
}
