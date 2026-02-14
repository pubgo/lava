package process

import (
	"debug/buildinfo"

	"github.com/gofiber/fiber/v3"
	ps "github.com/keybase/go-ps"
	"github.com/pubgo/funk/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/debug"
)

func init() {
	debug.Get("/process", func(ctx fiber.Ctx) (gErr error) {
		defer result.RecoveryErr(&gErr)
		processes := assert.Must1(ps.Processes())
		processes1 := funk.Map(processes, func(p ps.Process) map[string]any {
			path, err := p.Path()
			if err != nil {
				log.Err(err).Str("path", p.Executable()).Msg("process path error")
			}

			return map[string]any{
				"pid":        p.Pid(),
				"ppid":       p.PPid(),
				"exec":       p.Executable(),
				"path":       path,
				"go_version": goVersion(path),
			}
		})
		processes1 = funk.Filter(processes1, func(m map[string]any) bool { return m != nil })

		return ctx.JSON(processes1)
	})
}

func goVersion(path string) string {
	if path == "" {
		return ""
	}

	info, err := buildinfo.ReadFile(path)
	if err != nil {
		log.Err(err).CallerSkipFrame(1).Str("path", path).Msg("goVersion error")
		return ""
	}

	return info.GoVersion
}
