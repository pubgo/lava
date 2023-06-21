package logging

type LogConfigLoader struct {
	Log *Config `yaml:"logger"`
}

type Config struct {
	Level  string `yaml:"level"`
	AsJson bool   `yaml:"as-json"`
}
