package logging

type LogConfigLoader struct {
	Log *Config `yaml:"logger"`
}

type Config struct {
	Level          string   `yaml:"level"`
	AsJson         bool     `yaml:"as_json"`
	DisableLoggers []string `yaml:"disable_loggers"`
}
