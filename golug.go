package golug

import (
	"github.com/pubgo/golug/golug_abc"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/xerror"
)

func Init() (err error) {
	defer xerror.RespErr(&err)

	// 初始化配置文件
	xerror.Panic(golug_config.Init())
	return nil
}

func Run(entries ...golug_abc.Entry) (err error) {
	defer xerror.RespErr(&err)
	xerror.Panic(golug_app.Run(entries...))
	return nil
}

func NewEntry(name string) golug_abc.Entry {
	return golug_entry.New(name)
}
