package fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Cfg struct {
	Prefork       bool   `json:"prefork"`
	ServerHeader  string `json:"server_header"`
	StrictRouting bool   `json:"strict_routing"`
	CaseSensitive bool   `json:"case_sensitive"`
	Immutable     bool   `json:"immutable"`
	UnescapePath  bool   `json:"unescape_path"`
	ETag          bool   `json:"etag"`
	BodyLimit     int    `json:"body_limit"`
	Concurrency   int    `json:"concurrency"`
	Templates     struct {
		Dir string `json:"dir"`
		Ext string `json:"ext"`
	} `json:"templates"`
	ReadTimeout               time.Duration `json:"read_timeout"`
	WriteTimeout              time.Duration `json:"write_timeout"`
	IdleTimeout               time.Duration `json:"idle_timeout"`
	ReadBufferSize            int           `json:"read_buffer_size"`
	WriteBufferSize           int           `json:"write_buffer_size"`
	CompressedFileSuffix      string        `json:"compressed_file_suffix"`
	ProxyHeader               string        `json:"proxy_header"`
	GETOnly                   bool          `json:"get_only"`
	DisableKeepalive          bool          `json:"disable_keepalive"`
	DisableDefaultDate        bool          `json:"disable_default_date"`
	DisableDefaultContentType bool          `json:"disable_default_content_type"`
	DisableHeaderNormalizing  bool          `json:"disable_header_normalizing"`
	DisableStartupMessage     bool          `json:"disable_startup_message"`
	ReduceMemoryUsage         bool          `json:"reduce_memory_usage"`
	Websocket                 *WsCfg        `json:"websocket" yaml:"websocket"`
}

func DefaultCfg() fiber.Config {
	return fiber.New().Config()
}
