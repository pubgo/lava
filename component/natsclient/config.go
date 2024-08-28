package natsclient

import "github.com/aginetwork7/portal-server/internal/component/gopool"

type Config struct {
	Url          string         `yaml:"url"`
	Pool         *gopool.Config `yaml:"pool"`
	EnableIgnore bool           `yaml:"enable_ignore"`
}
