package golug_rest

import "github.com/gofiber/fiber/v2"

const Name = "rest_entry"

type Cfg = fiber.Config

//xerror.Panic(ent.UnWrap(func(entry http_entry.Entry) {
//	entry.Use(logger.New(logger.Config{Format: "${pid} - ${time} ${status} - ${latency} ${method} ${path}\n"}))
//}))
