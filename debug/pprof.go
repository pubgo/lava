package debug

import (
	"errors"
	"net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/gofiber/fiber/v2"
)

func init() {
	Get("/gprof", Wrap(fgprof.Handler()))
	Route("/pprof", func(r fiber.Router) {
		r.Get("/", WrapFunc(pprof.Index))
		r.Get("/:name", func(ctx *fiber.Ctx) error {
			var name = ctx.Params("name")
			switch name {
			case "cmdline":
				return WrapFunc(pprof.Cmdline)(ctx)
			case "profile":
				return WrapFunc(pprof.Profile)(ctx)
			case "symbol":
				return WrapFunc(pprof.Symbol)(ctx)
			case "trace":
				return WrapFunc(pprof.Trace)(ctx)
			case "allocs":
				return Wrap(pprof.Handler("allocs"))(ctx)
			case "goroutine":
				return Wrap(pprof.Handler("goroutine"))(ctx)
			case "heap":
				return Wrap(pprof.Handler("heap"))(ctx)
			case "mutex":
				return Wrap(pprof.Handler("mutex"))(ctx)
			case "threadcreate":
				return Wrap(pprof.Handler("threadcreate"))(ctx)
			}

			return errors.New("name not found")
		})
	})
}
