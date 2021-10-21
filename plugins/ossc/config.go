package ossc

import (
	"github.com/pubgo/lava/plugins/logger"
	"go.uber.org/zap"
)

var Name = "oss"
var cfgList = make(map[string]ClientCfg)
var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named(Name) })
}

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
