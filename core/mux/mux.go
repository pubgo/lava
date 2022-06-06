package mux

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/dix"
)

func Invoke(fn func(mux *Mux)) { dix.Register(fn) }

func init() {
	dix.Register(func() *Mux {
		return &Mux{App: fiber.New()}
	})
}

type Mux struct {
	*fiber.App
}
