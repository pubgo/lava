package logging

type Config struct {
	Level  string `yaml:"level"`
	AsJson bool   `yaml:"as-json"`
}
