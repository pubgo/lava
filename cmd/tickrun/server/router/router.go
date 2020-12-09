package router

import "github.com/gofiber/fiber/v2"

func Api(r fiber.Router) {
	admin := r.Group("/admin")
	_ = admin


	api := r.Group("/api")
	_ = api
}
