package broker

var Name = "broker"
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Driver string `json:"driver"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "nsq",
	}
}
