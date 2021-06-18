package log

import (
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/xlog/xlog_config"
)

var name = "log"
var cfg = xlog_config.NewProdConfig()

func init() {
	if runenv.IsDev() || runenv.IsTest() {
		cfg = xlog_config.NewDevConfig()
	}
}
