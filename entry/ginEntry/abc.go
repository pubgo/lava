package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/entry"
)

type Handler interface {
	entry.InitHandler
	Router(r gin.IRouter)
}

type Entry interface {
	entry.Entry
	Register(handler Handler)
}
