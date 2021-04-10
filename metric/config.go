package metric

var Name = "metric"

type Cfg struct {
	Driver string `json:"driver"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "noop",
	}
}
