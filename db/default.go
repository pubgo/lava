package db

import (
	"github.com/pubgo/xlog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"

	"github.com/pubgo/dix"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/x/pathutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	xl "xorm.io/xorm/log"
	"xorm.io/xorm/names"
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

	source := config.Template(cfg.Source)
	if strings.Contains(cfg.Driver, "sqlite") {
		if _dir := filepath.Dir(source); pathutil.IsNotExist(_dir) {
			_ = os.MkdirAll(_dir, 0755)
		}
	}

	engine := xerror.PanicErr(xorm.NewEngine(cfg.Driver, source)).(*xorm.Engine)
	engine.SetMaxOpenConns(cfg.MaxConnOpen)
	engine.SetMaxIdleConns(cfg.MaxConnIdle)
	engine.SetConnMaxLifetime(cfg.MaxConnTime)
	engine.SetMapper(names.LintGonicMapper)
	engine.Logger().SetLevel(xl.LOG_WARNING)
	if cfg.Debug && (config.IsDev() || config.IsTest()) {
		engine.Logger().SetLevel(xl.LOG_DEBUG)
		engine.ShowSQL(true)
	}

	xerror.Panic(engine.DB().Ping())

	val, ok := clients.Load(name)

	var client = &Client{engine}
	runtime.SetFinalizer(client, func(c *Client) {
		xlog.Infof("old orm client %s object %d gc", uintptr(unsafe.Pointer(c)))
		if err := c.Close(); err != nil {
			xlog.Errorf("orm close error, name: %s, err:%#v", err)
		}
	})

	clients.Set(name, client)

	// TODO runtime.SetFinalizer() 处理下
	if ok {
		_ = val.(*Client).Close()
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
