package fiber_builder

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/pubgo/funk/generic"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
)

type Config struct {
	Prefork       bool   `yaml:"prefork"`
	ServerHeader  string `yaml:"server_header"`
	StrictRouting bool   `yaml:"strict_routing"`
	CaseSensitive bool   `yaml:"case_sensitive"`
	Immutable     bool   `yaml:"immutable"`
	UnescapePath  bool   `yaml:"unescape_path"`
	ETag          bool   `yaml:"etag"`
	BodyLimit     int    `yaml:"body_limit"`
	Concurrency   int    `yaml:"concurrency"`
	Templates     struct {
		Dir string `yaml:"dir"`
		Ext string `yaml:"ext"`
	} `yaml:"templates"`
	ReadTimeout               time.Duration `yaml:"read_timeout"`
	WriteTimeout              time.Duration `yaml:"write_timeout"`
	IdleTimeout               time.Duration `yaml:"idle_timeout"`
	ReadBufferSize            int           `yaml:"read_buffer_size"`
	WriteBufferSize           int           `yaml:"write_buffer_size"`
	CompressedFileSuffix      string        `yaml:"compressed_file_suffix"`
	ProxyHeader               string        `yaml:"proxy_header"`
	GETOnly                   bool          `yaml:"get_only"`
	DisableKeepalive          bool          `yaml:"disable_keepalive"`
	DisableDefaultDate        bool          `yaml:"disable_default_date"`
	DisableDefaultContentType bool          `yaml:"disable_default_content_type"`
	DisableHeaderNormalizing  bool          `yaml:"disable_header_normalizing"`
	DisableStartupMessage     bool          `yaml:"disable_startup_message"`
	ReduceMemoryUsage         bool          `yaml:"reduce_memory_usage"`
}

func (t *Config) Build() (r result.Result[*fiber.App]) {
	defer recovery.Result(&r)
	fc := merge.Struct(generic.Ptr(fiber.New().Config()), &t).Unwrap()
	if t.Templates.Dir != "" && t.Templates.Ext != "" {
		fc.Views = html.New(t.Templates.Dir, t.Templates.Ext)
	}
	return r.WithVal(fiber.New(*fc))
}
