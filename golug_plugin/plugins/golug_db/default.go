package golug_db

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	sqlite "github.com/mattn/go-sqlite3"
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

func GetClient(names ...string) (*xorm.Engine, error) {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}

	val, ok := clientM.Load(name)
	if !ok {
		return nil, xerror.Fmt("%s not found", name)
	}

	return val.(*xorm.Engine), nil
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

var engine *xorm.Engine

func initClient(name string, cfg ClientCfg) {
	engine = xerror.PanicErr(xorm.NewEngine(cfg.Driver, cfg.Source)).(*xorm.Engine)
	if golug_env.IsDev() {
		engine.ShowSQL(true)
		engine.Logger().SetLevel(xl.LOG_DEBUG)
	} else {
		engine.ShowSQL(false)
		engine.Logger().SetLevel(xl.LOG_WARNING)
	}

	xerror.Panic(engine.DB().Ping())

	engine.SetMapper(names.GonicMapper{})

	fmt.Printf("%#v\n", xerror.PanicErr(engine.DBMetas()))

	if golug_env.IsDev() {
		xerror.Panic(engine.DumpAll(os.Stdout))
	}

	clientM.Store(name, engine)
}
