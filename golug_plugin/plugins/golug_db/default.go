package golug_db

import (
	"database/sql"
	"github.com/pubgo/golug/pkg/golug_util"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	sqlite "github.com/mattn/go-sqlite3"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/dialects"
	xl "xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

func GetCfg() Cfg {
	return cfg
}

func GetClient(names ...string) *xorm.Engine {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", name))
	}

	return val.(*xorm.Engine)
}

func floor(x float64) float64 {
	return math.Floor(x)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random() float32 {
	return rand.Float32()
}

func init() {
	sql.Register("sqlite3_custom", &sqlite.SQLiteDriver{
		ConnectHook: func(conn *sqlite.SQLiteConn) error {
			if err := conn.RegisterFunc("floor", floor, true); err != nil {
				return err
			}

			if err := conn.RegisterFunc("rand", random, true); err != nil {
				return err
			}
			return nil
		},
	})
	dialects.RegisterDriver("sqlite3_custom", dialects.QueryDriver("sqlite3"))
	dialects.RegisterDialect("sqlite3_custom", func() dialects.Dialect { return dialects.QueryDialect("sqlite3") })
}

func initClient(name string, cfg ClientCfg) {
	source := golug_config.Fmt(cfg.Source)
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
	engine.SetMapper(names.GonicMapper{})
	clientM.Store(name, engine)
}
