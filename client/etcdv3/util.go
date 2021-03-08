package etcdv3

import (
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/xerror"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// cfgMerge 合并etcd Cfg
func cfgMerge(cfg clientv3.Config) (cfg1 clientv3.Config, err error) {
	cfg1 = GetDefaultCfg().ToEtcdConfig()
	err = xerror.WrapF(gutils.Mergo(&cfg1, cfg), "[etcd] client Cfg merge error")
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
