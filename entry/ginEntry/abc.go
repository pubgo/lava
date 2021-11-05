package ginEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/entry"
)

type Router interface {
	Router(r gin.IRouter)
}

type Entry interface {
	entry.Entry
	Register(entry.InitHandler)
}
