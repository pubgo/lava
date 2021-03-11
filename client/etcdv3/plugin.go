package etcdv3

import (
	"strings"

	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	config.Decode(Name, &cfgList)

	for name, cfg := range cfgList {
		// etcd config处理
		cfg := xerror.PanicErr(cfgMerge(cfg)).(Cfg)
		xerror.Panic(initClient(consts.GetDefault(name), cfg))
	}
}

func onWatch(r *watcher.Response) {
	keys := strings.Split(r.Key, "/")
	name := keys[len(keys)-1]

	r.OnPut(func() {
		log.Debugf("[etcd] update client %s", name)

		// 解析etcd配置
		var cfg Cfg
		xerror.PanicF(r.Decode(&cfg), "[etcd] clientv3 Config parse error, cfgList: %s", r.Value)

		cfg = xerror.PanicErr(cfgMerge(cfg)).(Cfg)
		xerror.PanicF(updateClient(name, cfg), "[etcd] client %s watcher update error", name)
	})

	r.OnDelete(func() {
		log.Debugf("[etcd] delete client %s", name)

		if Get(name) == nil {
			log.Errorf("[etcd] client %s not found", name)
		}

		delClient(name)
	})
}

func init() {
	plugin.Register(&plugin.Base{
		Name:    Name,
		OnInit:  onInit,
		OnWatch: onWatch,
	})
}
