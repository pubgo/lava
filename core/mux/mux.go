package mux

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

func init() {
	dix.Register(func() *Mux {
		return &Mux{App: fiber.New()}
	})
}

type Mux struct {
	*fiber.App
}
