package schedulerdebug

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/scheduler"
)

func Init(scheduler scheduler.JobManager) {
	debug.Route("/scheduler", func(router fiber.Router) {
		router.Get("list", func(ctx *fiber.Ctx) error {
			ctx.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
			return Page(time.Now(), scheduler.ListJobs()).Render(ctx)
		})
	})
}
