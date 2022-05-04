package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	fx.In
	L *zap.Logger
}

func (t *Handler) Router(r fiber.Router) {
	r.Get("/hello", t.Get)
}

func (t *Handler) Get(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"hello": "ok"})
}
