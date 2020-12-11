package golug_entry_task

type Cfg struct {
	Enabled   bool   `yaml:"enabled" json:"enabled" toml:"enabled"`
	Broker    string `yaml:"broker"`
	Consumers []struct {
		Driver string `json:"driver" yaml:"driver"`
		Name   string `json:"name" yaml:"name"`
	} `json:"consumers" yaml:"consumers"`
}
