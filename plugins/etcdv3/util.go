package etcdv3

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

// cfgMerge 合并etcd Cfg
func cfgMerge(cfg Cfg) (cfg1 Cfg, err error) {
	cfg1 = GetDefaultCfg()
	err = xerror.WrapF(merge.Copy(&cfg1, &cfg), "config merge error")
	return
}
