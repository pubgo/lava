package gin

import "github.com/gin-gonic/gin"

type Builder struct {
	srv *gin.Engine
}

func (t *Builder) Get() *gin.Engine {
	if t.srv == nil {
		panic("please init gin")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	return engine
}
