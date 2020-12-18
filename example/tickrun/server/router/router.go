package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/example/tickrun/server/router/api/task"
	"github.com/pubgo/xerror"
)

func Api(r fiber.Router) {
	r.Use(func(view *fiber.Ctx) error {
		defer xerror.Resp(func(err xerror.XErr) {
			_ = view.JSON(fiber.Map{
				"code":   400,
				"detail": err,
				"msg":    err.Error(),
			})
		})

		return view.Next()
	})

	admin := r.Group("/admin")
	_ = admin

	api := r.Group("/api")
	api.Post("/task", task.Create)
	api.Delete("/task/:id", task.Delete)
	api.Put("/task/:id", task.Update)
	api.Get("/task/:id", task.Find)
	api.Get("/tasks", task.List)
	_ = api
}
