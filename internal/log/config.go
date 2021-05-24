package log

import (
	"github.com/pubgo/lug/app"
	"github.com/pubgo/xlog/xlog_config"
)

var name = "log"
var cfg = xlog_config.NewProdConfig()

func init() {
	if app.IsDev() || app.IsTest() {
		cfg = xlog_config.NewDevConfig()
	}
}
