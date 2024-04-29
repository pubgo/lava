package process

import (
	"debug/buildinfo"

	"github.com/gofiber/fiber/v2"
	ps "github.com/keybase/go-ps"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/result"

	"github.com/pubgo/lava/core/debug"
)

func init() {
	debug.Get("/process", func(ctx *fiber.Ctx) error {
		processes := assert.Must1(ps.Processes())
		processes1 := generic.Map(processes, func(i int) map[string]any {
			p := processes[i]
			ret := goVersion(result.Wrap(p.Path()))
			if ret.IsErr() {
				return nil
			}

			return map[string]any{
				"pid":        p.Pid(),
				"ppid":       p.PPid(),
				"exec":       p.Executable(),
				"path":       result.Wrap(p.Path()),
				"go_version": ret.Unwrap(),
			}
		})
		processes1 = generic.Filter(processes1, func(m map[string]any) bool { return m != nil })

		return ctx.JSON(processes1)
	})
}

func goVersion(path result.Result[string]) result.Result[string] {
	if path.IsErr() {
		return path
	}

	info, err := buildinfo.ReadFile(path.Unwrap())
	if err != nil {
		return result.Wrap("", err)
	}

	return result.OK(info.GoVersion)
}
