package golug_db

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	xl "xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

var clientM sync.Map

func GetClient(names ...string) *xorm.Engine {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("[db] %s not found", name))
	}

	return val.(*xorm.Engine)
}

func initClient(name string, cfg ClientCfg) {
	source := golug_config.Template(cfg.Source)
	if strings.Contains(cfg.Driver, "sqlite") {
		if _dir := filepath.Dir(source); !golug_util.PathExist(_dir) {
			_ = os.MkdirAll(_dir, 0755)
		}
	}

	engine := xerror.PanicErr(xorm.NewEngine(cfg.Driver, source)).(*xorm.Engine)
	engine.Logger().SetLevel(xl.LOG_WARNING)
	if golug_env.IsDev() {
		engine.Logger().SetLevel(xl.LOG_DEBUG)
	}
	xerror.Panic(engine.DB().Ping())
	engine.SetMapper(names.LintGonicMapper)
	clientM.Store(name, engine)
}
