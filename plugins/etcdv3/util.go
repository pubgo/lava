package etcdv3

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

// cfgMerge 合并etcd Cfg
func cfgMerge(cfg Cfg) (cfg1 Cfg, err error) {
	cfg1 = GetDefaultCfg()
	err = xerror.WrapF(merge.Copy(&cfg1, &cfg), "config merge error")
	return
}

func retry(c int, fn func() error) (err error) {
	for i := 0; i < c; i++ {
		if err = fn(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	return
}
