package golug_oss

var Name = "oss"
var cfg = make(map[string]ClientCfg)

type ClientCfg struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
}

func GetCfg() map[string]ClientCfg {
	return cfg
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{}
}
