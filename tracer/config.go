package tracer

var Name = "tracer"

type Cfg struct {
	Driver string `json:"driver"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "jaeger",
	}
}
