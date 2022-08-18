package pprof

import (
	"errors"
	"net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lava/debug"
)

func init() {
	debug.Get("/gprof", debug.Wrap(fgprof.Handler()))
	debug.Route("/pprof", func(r fiber.Router) {
		r.Get("/", debug.WrapFunc(pprof.Index))
		r.Get("/:name", func(ctx *fiber.Ctx) error {
			var name = ctx.Params("name")
			switch name {
			case "cmdline":
				return debug.WrapFunc(pprof.Cmdline)(ctx)
			case "profile":
				return debug.WrapFunc(pprof.Profile)(ctx)
			case "symbol":
				return debug.WrapFunc(pprof.Symbol)(ctx)
			case "trace":
				return debug.WrapFunc(pprof.Trace)(ctx)
			case "allocs":
				return debug.Wrap(pprof.Handler("allocs"))(ctx)
			case "goroutine":
				return debug.Wrap(pprof.Handler("goroutine"))(ctx)
			case "heap":
				return debug.Wrap(pprof.Handler("heap"))(ctx)
			case "mutex":
				return debug.Wrap(pprof.Handler("mutex"))(ctx)
			case "threadcreate":
				return debug.Wrap(pprof.Handler("threadcreate"))(ctx)
			}
			return errors.New("name not found")
		})
	})
}
