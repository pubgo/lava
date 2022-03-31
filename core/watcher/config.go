package watcher_config

type Cfg struct {
	Driver    string                 `json:"driver"`
	DriverCfg map[string]interface{} `json:"driver_config"`
	SkipNull  bool                   `json:"skip_null"`

	// Projects 需要watcher的项目
	Projects []string `json:"projects"`
}

func DefaultCfg() Cfg {
	return Cfg{
		SkipNull: true,
	}
}
