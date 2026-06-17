package zrpcc

import "time"

type Config struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

func DefaultCfg() *Config {
	return &Config{
		URL:     "nats://127.0.0.1:4222",
		Timeout: 3 * time.Second,
	}
}
