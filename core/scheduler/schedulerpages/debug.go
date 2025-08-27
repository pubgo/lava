package schedulerpages

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/scheduler"
)

func Init(scheduler scheduler.JobManager) {
	debug.Route("/scheduler", func(router fiber.Router) {
		router.Get("list", func(ctx *fiber.Ctx) error {
			scheduler.ListJobs()
			ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
			return Page(time.Now()).Render(ctx)
		})
	})
}
