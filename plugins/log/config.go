package log

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/xlog/xlog_config"
)

var name = "log"
var cfg = xlog_config.NewProdConfig()

func init() {
	if config.IsDev() || config.IsTest() {
		cfg = xlog_config.NewDevConfig()
	}
}
