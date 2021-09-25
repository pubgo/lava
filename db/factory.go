package db

import (
	"runtime"
	"unsafe"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/atomic"
	"xorm.io/xorm"
)

var clients typex.SMap

type Client struct {
	val atomic.Value
}

func (c *Client) Get() *xorm.Engine {
	return c.val.Load().(*xorm.Engine)
}

func Get(names ...string) *xorm.Engine {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client).Get()
}

func GetCallback(cb func(*xorm.Engine), names ...string) {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return
	}

	cb(c.(*Client).Get())
}

func Update(name string, cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	var client = &Client{}
	client.val.Store(xerror.PanicErr(cfg.Build()).(*xorm.Engine))

	val, ok := clients.Load(name)
	if ok {
		val.(*Client).val.Store(client)
		return
	}

	runtime.SetFinalizer(client, func(c *Client) {
		logs.Sugar().Infof("old db client %s object %d gc", name, uintptr(unsafe.Pointer(c)))
		if err := c.Get().Close(); err != nil {
			logs.Sugar().Errorw("db close error", "name", name, logger.Err(err))
		}
	})

	clients.Set(name, client)
	// 初始化完毕之后, 更新到对象管理系统
	xerror.Panic(dix.Provider(map[string]*Client{name: client}))

	return
}

func Delete(name string) {
	clients.Delete(consts.GetDefault(name))
}
