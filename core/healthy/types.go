package healthy

import (
	"github.com/gofiber/fiber/v3"
)

// Handler 是单个健康检查的处理函数，返回 error 表示不健康。
type Handler func(req fiber.Ctx) error
