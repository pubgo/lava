package golug_db_mysql

import (
	"database/sql"
	"math"
	"math/rand"
	"time"

	sqlite "github.com/mattn/go-sqlite3"
	"xorm.io/xorm/dialects"
)

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
	sql.Register("sqlite3_golug", &sqlite.SQLiteDriver{
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
	dialects.RegisterDriver("sqlite3_golug", dialects.QueryDriver("sqlite3"))
	dialects.RegisterDialect("sqlite3_golug", func() dialects.Dialect { return dialects.QueryDialect("sqlite3") })
}
