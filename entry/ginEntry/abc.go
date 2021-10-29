package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/entry"
)

type Handler interface {
	entry.InitHandler
	Group(r *gin.RouterGroup)
}

type Entry interface {
	entry.Entry
	Register(handler Handler)
}
