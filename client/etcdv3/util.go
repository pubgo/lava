package etcdv3

import (
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
)

// cfgMerge 合并etcd config
func cfgMerge(cfg clientv3.Config) (cfg1 clientv3.Config, err error) {
	cfg1 = DefaultCfg
	if err1 := mergo.Map(&cfg1, cfg, mergo.WithOverride, mergo.WithTypeCheck); err1 != nil {
		err = errors.Wrapf(err1, "[etcd] client config merge error")
	}
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
