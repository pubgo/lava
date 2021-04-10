package fiber

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/internal/utils"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func Start(app *fiber.App, port int) (err error) {
	defer xerror.RespErr(&err)

	if port < 1 {
		port, _ = utils.GetFreePort()
	}

	// 启动server后等待1s
	xerror.Panic(fx.GoDelay(time.Second, func() {
		defer xerror.Resp(func(err xerror.XErr) {
			xlog.Error("app start error", xlog.Any("err", err))
		})

		for {
			if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
				if strings.Contains(err.Error(), "address already in use") {
					port += 1
					continue
				}

				xlog.Error(xerror.Parse(err).Stack(true))
			}
			break
		}

		xlog.Infof("Server [http] Closed OK")
	}))
	xlog.Infof("Server [http] Listening on http://%d", port)

	return nil
}

func Stop(app *fiber.App) (err error) {
	return xutil.Try(func() {
		if err := app.Shutdown(); err != nil && err != http.ErrServerClosed {
			xlog.Error(xerror.Parse(err).Stack(true))
		}
	})
}
