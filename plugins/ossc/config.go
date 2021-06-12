package ossc

var Name = "oss"
var cfgList = make(map[string]ClientCfg)

type ClientCfg struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
}

func GetCfg() map[string]ClientCfg {
	return cfgList
}

func GetDefaultCfg() ClientCfg {
	return ClientCfg{}
}
