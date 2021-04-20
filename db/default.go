package db

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"xorm.io/xorm"

	"runtime"
	"unsafe"
)

var clients types.SMap

func Get(names ...string) *Client {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client)
}

func updateClient(name string, cfg Cfg) (err error) {
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
	return nil
}

func updateEngine(name string, engine *xorm.Engine) {
	xerror.Panic(dix.Dix(map[string]*xorm.Engine{name: engine}))
}

func Watch(db interface{}) {
	xerror.Panic(dix.Dix(db))
}

func delClient(name string) {
	var client = Get(name)
	if client == nil {
		return
	}

	runtime.SetFinalizer(client, func(c *Client) {
		xlog.Infof("old db client %s object %d gc", name, uintptr(unsafe.Pointer(c)))
		if err := c.Close(); err != nil {
			xlog.Errorf("db close error, name: %s", name, xlog.Any("err", err))
		}
	})
	clients.Delete(name)
}
