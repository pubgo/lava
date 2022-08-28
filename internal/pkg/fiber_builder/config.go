package fiber_builder

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/lava/internal/pkg/merge"
	"time"
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

func (t *Cfg) Build() (ret result.Result[*fiber.App]) {
	defer recovery.Result(&ret)

	var fc = fiber.New().Config()
	assert.Must(merge.Struct(&fc, &t))
	if t.Templates.Dir != "" && t.Templates.Ext != "" {
		fc.Views = html.New(t.Templates.Dir, t.Templates.Ext)
	}

	return ret.WithVal(fiber.New(fc))
}
