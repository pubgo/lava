package fiberbuilder

import (
	"log/slog"
	"time"

	"dario.cat/mergo"
	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"

	"github.com/pubgo/lava/v2/pkg/encoding/protojson"
)

type Config struct {
	ServerHeader              string             `yaml:"-"`
	CaseSensitive             bool               `yaml:"-"`
	Immutable                 bool               `yaml:"-"`
	UnescapePath              bool               `yaml:"-"`
	ProxyHeader               string             `yaml:"-"`
	GETOnly                   bool               `yaml:"-"`
	DisableKeepalive          bool               `yaml:"-"`
	DisableDefaultDate        bool               `yaml:"-"`
	DisableDefaultContentType bool               `yaml:"-"`
	ErrorHandler              fiber.ErrorHandler `yaml:"-" json:"-"`
	BodyLimit                 int                `yaml:"-"`
	Concurrency               int                `yaml:"-"`

	// StreamRequestBody enables request body streaming,
	// and calls the handler sooner when given body is
	// larger then the current limit.
	StreamRequestBody bool `yaml:"stream_request_body"`

	// Will not pre parse Multipart Form data if set to true.
	//
	// This option is useful for servers that desire to treat
	// multipart form data as a binary blob, or choose when to parse the data.
	//
	// Server pre parses multipart form data by default.
	DisablePreParseMultipartForm bool `yaml:"disable_pre_parse_multipart_form"`

	// Aggressively reduces memory usage at the cost of higher CPU usage
	// if set to true.
	//
	// Try enabling this option only if the server consumes too much memory
	// serving mostly idle keep-alive connections. This may reduce memory
	// usage by more than 50%.
	//
	// Default: false
	ReduceMemoryUsage bool `yaml:"reduce_memory_usage"`

	// If set to true, c.IP() and c.IPs() will validate IP addresses before returning them.
	// Also, c.IP() will return only the first valid IP rather than just the raw header
	// WARNING: this has a performance cost associated with it.
	//
	// Default: false
	EnableIPValidation bool `yaml:"enable_ip_validation"`

	// If set to true, will print all routes with their method, path and handler.
	// Default: false
	EnablePrintRoutes bool `yaml:"enable_print_routes"`

	// EnableSplittingOnParsers splits the query/body/header parameters by comma when it's true.
	// For example, you can use it to parse multiple values from a query parameter like this:
	//   /api?foo=bar,baz == foo[]=bar&foo[]=baz
	//
	// Optional. Default: false
	EnableSplittingOnParsers bool `yaml:"enable_splitting_on_parsers"`

	ETag                     bool              `yaml:"etag"`
	ReadTimeout              time.Duration     `yaml:"read_timeout"`
	WriteTimeout             time.Duration     `yaml:"write_timeout"`
	IdleTimeout              time.Duration     `yaml:"idle_timeout"`
	ReadBufferSize           int               `yaml:"read_buffer_size"`
	WriteBufferSize          int               `yaml:"write_buffer_size"`
	CompressedFileSuffix     string            `yaml:"compressed_file_suffix"`
	CompressedFileSuffixes   map[string]string `yaml:"compressed_file_suffixes"`
	DisableHeaderNormalizing bool              `yaml:"disable_header_normalizing"`
}

func (t *Config) ToCfg() fiber.Config {
	compressed := t.CompressedFileSuffixes
	if compressed == nil && t.CompressedFileSuffix != "" {
		compressed = map[string]string{
			"gzip": t.CompressedFileSuffix,
		}
	}

	return fiber.Config{
		ServerHeader:                 t.ServerHeader,
		CaseSensitive:                t.CaseSensitive,
		Immutable:                    t.Immutable,
		UnescapePath:                 t.UnescapePath,
		ProxyHeader:                  t.ProxyHeader,
		GETOnly:                      t.GETOnly,
		DisableKeepalive:             t.DisableKeepalive,
		DisableDefaultDate:           t.DisableDefaultDate,
		DisableDefaultContentType:    t.DisableDefaultContentType,
		ErrorHandler:                 t.ErrorHandler,
		BodyLimit:                    t.BodyLimit,
		Concurrency:                  t.Concurrency,
		StreamRequestBody:            t.StreamRequestBody,
		DisablePreParseMultipartForm: t.DisablePreParseMultipartForm,
		ReduceMemoryUsage:            t.ReduceMemoryUsage,
		EnableIPValidation:           t.EnableIPValidation,
		EnableSplittingOnParsers:     t.EnableSplittingOnParsers,
		ReadTimeout:                  t.ReadTimeout,
		WriteTimeout:                 t.WriteTimeout,
		IdleTimeout:                  t.IdleTimeout,
		ReadBufferSize:               t.ReadBufferSize,
		WriteBufferSize:              t.WriteBufferSize,
		CompressedFileSuffixes:       compressed,
		DisableHeaderNormalizing:     t.DisableHeaderNormalizing,
		JSONEncoder:                  protojson.Default.Marshal,
	}
}

func (t *Config) Build() (r result.Result[fiber.Config]) {
	if t == nil {
		return r.WithValue(fiber.New().Config())
	}

	defer result.Recovery(&r)
	cfg := fiber.New().Config()
	err := mergo.Merge(&cfg, lo.ToPtr(t.ToCfg()), mergo.WithOverride, mergo.WithAppendSlice)
	if err != nil {
		slog.Error("failed to merge config", "err", err, "source", t, "target", cfg)
		return r.WithErrorf("failed to merge config, err:%v", err)
	}
	return r.WithValue(cfg)
}
