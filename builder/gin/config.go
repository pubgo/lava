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
