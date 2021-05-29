package gin

import "github.com/gin-gonic/gin"

func init() {
	gin.SetMode(gin.DebugMode)
}

type Cfg struct {
	DisableBindValidation                  bool
	EnableJsonDecoderUseNumber             bool
	EnableJsonDecoderDisallowUnknownFields bool
	Mode                                   string
}

func (t *Cfg) Build() (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())
	return engine, nil
}
