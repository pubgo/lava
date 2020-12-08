package golug_db

type Cfg struct {
	Configs map[string]ClientCfg `yaml:"configs" json:"configs" toml:"configs"`
}

type ClientCfg struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Driver  string `json:"driver" yaml:"driver"`
	Source  string `json:"source" yaml:"source"`
}
