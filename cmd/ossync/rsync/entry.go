package rsync

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	badger "github.com/dgraph-io/badger/v2"
	"github.com/pubgo/golug"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xprocess"
	"go.uber.org/atomic"
)

var name = "ossync"

func init() {
	golug_env.Project = name
}

func GetEntry() golug_entry.Entry {
	ent := golug.NewCtlEntry(name)
	xerror.Panic(ent.Version("v0.0.1"))
	xerror.Panic(ent.Description("sync from local to remote"))
	xerror.Exit(ent.Commands(GetDbCmd()))

	ent.Register(func() {
		defer xerror.RespDebug()

		client, err := oss.New(
			os.Getenv("oss_endpoint"),
			os.Getenv("oss_ak"),
			os.Getenv("oss_sk"),
		)
		xerror.Panic(err)
		kk := xerror.PanicErr(client.Bucket("kooksee")).(*oss.Bucket)

		opts := badger.DefaultOptions(filepath.Join(golug_env.Home, "db"))
		db, err := badger.Open(opts)
		xerror.Panic(err)
		defer db.Close()

		var nw = NewWaiter()
		var run = func(path string) {
			key := os.ExpandEnv(path)

			xprocess.GoLoop(func(ctx context.Context) {
				if nw.Skip(key) {
					time.Sleep(5 * time.Second)
					return
				}

				var c = atomic.NewBool(false)
				defer nw.Report(key, c)
				checkAndSync(key, kk, db, "", c)
				checkAndMove(kk, db, c)
				checkAndBackup(key, kk)
			})
		}

		run("${HOME}/Documents")
		run("${HOME}/Downloads")
		run("${HOME}/git/docs")
		select {}
	})

	return ent
}
