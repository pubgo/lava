package grpcEntry

import (
	"github.com/gin-gonic/gin"
	"github.com/pubgo/lava/entry"

	// grpc log插件加载
	_ "github.com/pubgo/lava/internal/plugins/grpclog"
)

type Handler interface {
	entry.InitHandler
	Group(r *gin.RouterGroup)
}

type Entry interface {
	entry.Entry
	Register(handler Handler)
}
