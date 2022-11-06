package https

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/service"
)

type Route struct {
	app *fiber.App
}

func (a *Route) Use(args ...interface{}) *Route {
	a.app.Use(args...)
	return a
}

func (a *Route) Get(path string, handler Handler[any, any], middlewares ...service.Middleware) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Head(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Post(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Put(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Delete(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Connect(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Options(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Trace(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Patch(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Add(method, path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) All(path string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Group(prefix string, handlers ...fiber.Handler) *Route {
	//TODO implement me
	panic("implement me")
}

func (a *Route) Route(prefix string, fn func(router *Route), name ...string) *Route {
	//TODO implement me
	panic("implement me")
}
