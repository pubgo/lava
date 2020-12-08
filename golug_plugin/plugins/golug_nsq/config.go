package golug_nsq

import "time"

type Cfg struct {
	Configs map[string]ClientCfg `yaml:"configs" json:"configs" toml:"configs"`
}

type ClientCfg struct {
	Enabled        bool          `yaml:"enabled" json:"enabled" toml:"enabled"`
	Topic          string        `json:"topic"`
	Channel        string        `json:"channel"`
	Address        string        `json:"address"`
	Lookup         []string      `json:"lookup"`
	MaxInFlight    int           `json:"max_in_flight"`
	MaxConcurrency int           `json:"max_concurrency"`
	DialTimeout    time.Duration `json:"dial_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	DrainTimeout   time.Duration `json:"drain_timeout"`
}
