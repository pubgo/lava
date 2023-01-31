package logconfig

import (
	"io"
)

type Config struct {
	Level  string `yaml:"level"`
	AsJson bool   `yaml:"as-json"`

	Writer io.Writer
}
