package zrpcs

type Config struct {
	URL string `yaml:"url"`
}

func defaultCfg() *Config {
	return &Config{URL: "nats://127.0.0.1:4222"}
}
