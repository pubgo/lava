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

func (t *Handler) App() *fiber.App {
	var app = fiber.New()
	return app
}

func (t *Handler) Get(ctx *fiber.Ctx) error {
	return nil
}
