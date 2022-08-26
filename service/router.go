package service

import "github.com/gofiber/fiber/v2"

type InitRouter interface {
	Router() *fiber.App
}
