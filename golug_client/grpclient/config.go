package grpclient

type Cfg struct {
	Configs map[string]ClientCfg `yaml:"configs" json:"configs" toml:"configs"`
}

type ClientCfg struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}
