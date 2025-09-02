package fiberbuilder

import (
	"github.com/samber/lo"
	"log/slog"
	"time"

	"dario.cat/mergo"
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/funk/v2/result"
)

type Config struct {
	Prefork                   bool               `yaml:"-"`
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

	ETag                     bool          `yaml:"etag"`
	ReadTimeout              time.Duration `yaml:"read_timeout"`
	WriteTimeout             time.Duration `yaml:"write_timeout"`
	IdleTimeout              time.Duration `yaml:"idle_timeout"`
	ReadBufferSize           int           `yaml:"read_buffer_size"`
	WriteBufferSize          int           `yaml:"write_buffer_size"`
	CompressedFileSuffix     string        `yaml:"compressed_file_suffix"`
	DisableHeaderNormalizing bool          `yaml:"disable_header_normalizing"`
	DisableStartupMessage    bool          `yaml:"disable_startup_message"`
}

func (t *Config) ToCfg() fiber.Config {
	return fiber.Config{
		Prefork:                      t.Prefork,
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
		EnablePrintRoutes:            t.EnablePrintRoutes,
		EnableSplittingOnParsers:     t.EnableSplittingOnParsers,
		ETag:                         t.ETag,
		ReadTimeout:                  t.ReadTimeout,
		WriteTimeout:                 t.WriteTimeout,
		IdleTimeout:                  t.IdleTimeout,
		ReadBufferSize:               t.ReadBufferSize,
		WriteBufferSize:              t.WriteBufferSize,
		CompressedFileSuffix:         t.CompressedFileSuffix,
		DisableHeaderNormalizing:     t.DisableHeaderNormalizing,
		DisableStartupMessage:        t.DisableStartupMessage,
	}
}

func (t *Config) Build() (r result.Result[fiber.Config]) {
	if t == nil {
		return r.WithValue(fiber.New().Config())
	}

	defer result.RecoveryErr(&r)
	cfg := fiber.New().Config()
	err := mergo.Merge(&cfg, lo.ToPtr(t.ToCfg()), mergo.WithOverride, mergo.WithAppendSlice)
	if err != nil {
		slog.Error("failed to merge config", "err", err, "source", t, "target", cfg)
		return r.WithErrorf("failed to merge config, err:%v", err)
	}
	return r.WithValue(cfg)
}
