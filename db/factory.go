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
	client atomic.Value
}

func (c *Client) Get() *xorm.Engine {
	return c.client.Load().(*xorm.Engine)
}

func Get(names ...string) *xorm.Engine {
	c := clients.Get(consts.GetDefault(names...))
	if c == nil {
		return nil
	}

	return c.(*Client).Get()
}

func GetWith(name string, cb func(*xorm.Engine)) {
	c := clients.Get(consts.GetDefault(name))
	if c == nil {
		return
	}

	cb(c.(*Client).Get())
}

func Update(name string, cfg Cfg) (err error) {
	defer xerror.RespErr(&err)

	var client = &Client{}
	client.client.Store(xerror.PanicErr(cfg.Build()).(*xorm.Engine))

	val, ok := clients.Load(name)
	// 更新client
	if ok && val != nil {
		val.(*Client).client.Store(client)
		return
	}

	// 创建client
	clients.Set(name, client)
	runtime.SetFinalizer(client, func(c *Client) {
		logs.Sugar().Infof("old db client %s object %d gc", name, uintptr(unsafe.Pointer(c)))
		if err := c.Get().Close(); err != nil {
			logs.Sugar().Errorw("db close error", "name", name, logger.Err(err))
		}
	})

	// 依赖注入
	xerror.Panic(dix.Provider(map[string]*Client{name: client}))

	return
}

func Delete(name string) {
	clients.Delete(consts.GetDefault(name))
}
