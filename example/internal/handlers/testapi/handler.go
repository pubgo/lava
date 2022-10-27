package testapi

import (
	"context"
	"fmt"

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
	app.Get("/hello", https.Handler(func(ctx context.Context, req *Req) (rsp *Rsp, err error) {
		return &Rsp{Data: "ok"}, nil
	}))

	app.Get("/error", https.Handler(func(ctx context.Context, req *Req) (rsp *Rsp, err error) {
		return nil, fmt.Errorf("this is error")
	}))
}

type Rsp struct {
	Data string `json:"data"`
}

type Req struct {
}
