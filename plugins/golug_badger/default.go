package golug_badger

import (
	"path/filepath"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

var clientM sync.Map

func GetClient(names ...string) *badger.DB {
	var name = golug_consts.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	val, ok := clientM.Load(name)
	if !ok {
		xlog.Warnf("%s not found", name)
		return nil
	}

	return val.(*badger.DB)
}

func initClient(name string, cfg ClientCfg) {
	dbPath := filepath.Join(golug_env.Home, cfg.Path)
	opts := badger.DefaultOptions(dbPath)
	opts.WithLoggingLevel(badger.DEBUG)

	db, err := badger.Open(opts)
	xerror.Panic(err)

	clientM.Store(name, db)
}
