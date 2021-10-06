package sqlite

import (
	"database/sql"
	"math"
	"math/rand"
	"time"

	sqlite "github.com/mattn/go-sqlite3"
	"xorm.io/xorm/dialects"
)

const Name = "sqlite3x"

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
	sql.Register(Name, &sqlite.SQLiteDriver{
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
	dialects.RegisterDriver(Name, dialects.QueryDriver("sqlite3"))
	dialects.RegisterDialect(Name, func() dialects.Dialect { return dialects.QueryDialect("sqlite3") })
}
