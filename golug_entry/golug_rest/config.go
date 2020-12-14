package golug_rest

import (
	"github.com/gofiber/fiber/v2"
)

const Name = "rest_entry"

var cfg = GetDefaultCfg()

func GetCfg() fiber.Config {
	return cfg
}

func GetDefaultCfg() fiber.Config {
	return fiber.New().Config()
}
