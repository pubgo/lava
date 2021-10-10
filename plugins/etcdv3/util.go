package etcdv3

import (
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

// cfgMerge 合并etcd Cfg
func cfgMerge(cfg Cfg) *Cfg {
	cfg1 := GetDefaultCfg()
	xerror.Panic(merge.Copy(cfg1, &cfg), "config merge error")
	return cfg1
}
