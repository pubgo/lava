package ossc

import (
	"github.com/pubgo/lava/logger"
)

var Name = "oss"
var cfgList = make(map[string]ClientCfg)
var logs = logger.Name(Name)

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
