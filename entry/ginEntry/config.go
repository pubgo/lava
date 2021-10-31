package ginEntry

import "github.com/gin-gonic/gin"

const Name = "gin_entry"

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Cfg struct {
	RedirectTrailingSlash  bool     `json:"redirect_trailing_slash" yaml:"redirect_trailing_slash"`
	RedirectFixedPath      bool     `json:"redirect_fixed_path" yaml:"redirect_fixed_path"`
	HandleMethodNotAllowed bool     `json:"handle_method_not_allowed" yaml:"handle_method_not_allowed"`
	ForwardedByClientIP    bool     `json:"forwarded_by_client_ip" yaml:"forwarded_by_client_ip"`
	RemoteIPHeaders        []string `json:"remote_ip_headers" yaml:"remote_ip_headers"`
	TrustedProxies         []string `json:"trusted_proxies" yaml:"trusted_proxies"`
	AppEngine              bool     `json:"app_engine" yaml:"app_engine"`
	UseRawPath             bool     `json:"use_raw_path" yaml:"use_raw_path"`
	UnescapePathValues     bool     `json:"unescape_path_values" yaml:"unescape_path_values"`
	MaxMultipartMemory     int64    `json:"max_multipart_memory" yaml:"max_multipart_memory"`
	RemoveExtraSlash       bool     `json:"remove_extra_slash" yaml:"remove_extra_slash"`
}
