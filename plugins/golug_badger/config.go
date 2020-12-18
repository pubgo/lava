package golug_badger

var Name = "badger"
var cfg = make(map[string]ClientCfg)

type ClientCfg struct {
	Path string `json:"path"`
}

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{}
}
