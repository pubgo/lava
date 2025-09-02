package tasks

import "github.com/pubgo/lava/v2/pkg/httputil"

type Config = httputil.Config

type HttpServerConfigLoader struct {
	HttpServer *Config `yaml:"http_server"`
}
