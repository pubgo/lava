package db

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"xorm.io/xorm"

	"runtime"
	"unsafe"
)

var clients typex.SMap

type Client struct {
	*xorm.Engine
}

func Get(names ...string) *Client {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client)
}

func Update(name string, cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	engine := xerror.PanicErr(cfg.Build()).(*xorm.Engine)

	val, ok := clients.Load(name)
	if !ok {
		clients.Set(name, &Client{engine})
	} else {
		val.(*Client).Engine = engine
	}

	// 初始化完毕之后, 更新到对象管理系统
	updateEngine(name, engine)
	return
}

func updateEngine(name string, engine *xorm.Engine) {
	xerror.Panic(dix.Dix(map[string]*xorm.Engine{name: engine}))
}

func Delete(name string) {
	var client = Get(name)
	if client == nil {
		return
	}

	runtime.SetFinalizer(client, func(c *Client) {
		logs.Infof("old db client %s object %d gc", name, uintptr(unsafe.Pointer(c)))
		if err := c.Close(); err != nil {
			logs.Errorf("db close error, name: %s", name, zap.Any("err", err))
		}
	})
	clients.Delete(name)
}
