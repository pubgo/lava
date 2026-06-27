package p2pdebug

import (
	"sync"

	"github.com/gofiber/fiber/v3"

	"github.com/pubgo/lava/v2/core/debug"
	"github.com/pubgo/lava/v2/core/p2p"
)

var (
	mu        sync.RWMutex
	coordinator p2p.Coordinator
)

func init() {
	debug.Get("/p2p", func(ctx fiber.Ctx) error {
		mu.RLock()
		c := coordinator
		mu.RUnlock()
		if c == nil {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "p2p not configured",
			})
		}
		return ctx.JSON(c.Stats())
	})
}

// Register 绑定 P2P 协调器；传 nil 表示注销。
func Register(c p2p.Coordinator) {
	mu.Lock()
	defer mu.Unlock()
	coordinator = c
}
