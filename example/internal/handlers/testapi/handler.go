package testapi

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/servers/https"
	"github.com/pubgo/lava/service"
)

func New() service.HttpRouter {
	return &handler{}
}

type handler struct {
}

func (h *handler) Init() {

}

func (h *handler) BasePrefix() string {
	return ""
}

func (h *handler) Middlewares() []service.Middleware {
	return nil
}

func (h *handler) Router(app *fiber.App) {
	app.Get("/hello", https.Wrap(func(ctx context.Context, req *Req) (rsp *Rsp, err error) {
		return &Rsp{Data: "ok"}, nil
	}))
}

type Rsp struct {
	Data string `json:"data"`
}

type Req struct {
}
