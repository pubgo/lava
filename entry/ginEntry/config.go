package ginEntry

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
