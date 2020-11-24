package task_entry

type Cfg struct {
	Enabled   bool `yaml:"enabled" json:"enabled" toml:"enabled"`
	Consumers []struct {
		Driver string `json:"driver"`
		Name   string `json:"name"`
	} `json:"consumers"`
}
