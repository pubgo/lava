package ratelimit

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/pubgo/lava/v2/core/debug"
)

var (
	enabled    = false
	limit      = 100
	window     = time.Minute
	requests   = make(map[string]*requestCounter)
	requestsMu sync.RWMutex
)

type requestCounter struct {
	Count     int
	ResetTime time.Time
}

func init() {
	debug.App().Use(rateLimitMiddleware)

	debug.Get("/ratelimit", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"enabled":        enabled,
			"limit":          limit,
			"window":         window.String(),
			"window_seconds": window.Seconds(),
		})
	})

	debug.Post("/ratelimit/enable", func(ctx *fiber.Ctx) error {
		enabled = true
		return ctx.JSON(fiber.Map{
			"success": true,
			"enabled": enabled,
		})
	})

	debug.Post("/ratelimit/disable", func(ctx *fiber.Ctx) error {
		enabled = false
		return ctx.JSON(fiber.Map{
			"success": true,
			"enabled": enabled,
		})
	})

	debug.Put("/ratelimit/config", func(ctx *fiber.Ctx) error {
		type request struct {
			Limit  int `json:"limit" form:"limit" query:"limit"`
			Window int `json:"window" form:"window" query:"window"`
		}

		var req request
		if err := ctx.BodyParser(&req); err != nil {
			if l := ctx.QueryInt("limit", 0); l > 0 {
				req.Limit = l
			}
			if w := ctx.QueryInt("window", 0); w > 0 {
				req.Window = w
			}
		}

		if req.Limit > 0 {
			limit = req.Limit
		}
		if req.Window > 0 {
			window = time.Duration(req.Window) * time.Second
		}

		return ctx.JSON(fiber.Map{
			"success":        true,
			"limit":          limit,
			"window":         window.String(),
			"window_seconds": window.Seconds(),
		})
	})

	debug.Get("/ratelimit/stats", func(ctx *fiber.Ctx) error {
		requestsMu.RLock()
		stats := make(map[string]fiber.Map)
		for ip, counter := range requests {
			stats[ip] = fiber.Map{
				"count":      counter.Count,
				"reset_time": counter.ResetTime.Format(time.RFC3339),
				"remaining":  limit - counter.Count,
			}
		}
		requestsMu.RUnlock()

		return ctx.JSON(fiber.Map{
			"enabled":      enabled,
			"limit":        limit,
			"window":       window.String(),
			"active_users": len(stats),
			"stats":        stats,
		})
	})

	debug.Post("/ratelimit/reset", func(ctx *fiber.Ctx) error {
		requestsMu.Lock()
		requests = make(map[string]*requestCounter)
		requestsMu.Unlock()

		return ctx.JSON(fiber.Map{
			"success": true,
			"message": "rate limit stats cleared",
		})
	})
}

func rateLimitMiddleware(ctx *fiber.Ctx) error {
	if !enabled {
		return ctx.Next()
	}

	ip := ctx.IP()

	requestsMu.Lock()
	counter, exists := requests[ip]
	if !exists || time.Now().After(counter.ResetTime) {
		counter = &requestCounter{
			Count:     0,
			ResetTime: time.Now().Add(window),
		}
		requests[ip] = counter
	}
	counter.Count++
	count := counter.Count
	resetTime := counter.ResetTime
	requestsMu.Unlock()

	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	ctx.Set("X-RateLimit-Limit", strconv.Itoa(limit))
	ctx.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	ctx.Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

	if count > limit {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":     "rate limit exceeded",
			"limit":     limit,
			"remaining": 0,
			"reset":     resetTime.Format(time.RFC3339),
		})
	}

	return ctx.Next()
}
