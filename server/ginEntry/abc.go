package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/server"
)

type Router interface {
	Router(r gin.IRouter)
}

type Entry interface {
	server.Entry
	Use(middleware ...gin.HandlerFunc)
	Register(server.Handler)
}
