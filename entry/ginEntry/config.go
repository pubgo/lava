package ginEntry

import "github.com/gin-gonic/gin"

const Name = "rest_entry"

type Cfg struct {
	RedirectTrailingSlash  bool
	RedirectFixedPath      bool
	HandleMethodNotAllowed bool
	ForwardedByClientIP    bool
	RemoteIPHeaders        []string
	TrustedProxies         []string
	AppEngine              bool
	UseRawPath             bool
	UnescapePathValues     bool
	MaxMultipartMemory     int64
	RemoveExtraSlash       bool
}

func (t Cfg) Build() *gin.Engine {
	var eng = gin.New()
	return eng
}
