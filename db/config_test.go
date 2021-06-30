package db

import (
	"fmt"
	"testing"

	"github.com/pubgo/lug/db/sqlite"
	"github.com/pubgo/xerror"
)

func TestConfig(t *testing.T) {
	defer xerror.RespTest(t)

	var cfg = GetDefaultCfg()
	cfg.Driver = sqlite.Name
	cfg.Source = "./sqlite.db"

	eng, err := cfg.Build()
	xerror.Panic(err)

	fmt.Println(eng.Query("select * from db"))
	fmt.Println(eng.Query("select * from db where Field1=?", 1))
}
