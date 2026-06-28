package tunnelcmd

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

const headerAdminToken = "X-Tunnel-Admin-Token"

// adminAuthMiddleware 保护 Admin UI；token 为空时不启用（开发模式）。
func adminAuthMiddleware(token string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if token == "" {
			return c.Next()
		}
		got := strings.TrimSpace(c.Get(headerAdminToken))
		if got == "" {
			auth := strings.TrimSpace(c.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				got = strings.TrimSpace(auth[7:])
			}
		}
		if got != token {
			return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
		}
		return c.Next()
	}
}
