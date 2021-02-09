package etcdv3

import (
	"strings"

	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/xerror"
)

func init() {
	golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
		},
		OnWatch: func(r *golug_watcher.Response) {
			r.OnPut(func() {
				keys := strings.Split(r.Key, "/")
				name := keys[len(keys)-1]

				log.Debugf("[etcd] update client %s", name)

				// 解析etcd配置
				var cfg config
				xerror.PanicF(r.Decode(&cfg), "[etcd] clientv3 Config parse error, cfg: %s", r.Value)
				xerror.PanicF(Update(name, cfg.EtcdConfig()), "[etcd] client %s watcher update error", name)
			})

			r.OnDelete(func() {
				keys := strings.Split(r.Key, "/")
				name := keys[len(keys)-1]

				log.Debugf("[etcd] delete client %s", name)
				if Get(name) == nil {
					log.Errorf("[etcd] client %s not found", name)
				}

				Del(name)
			})
		},
	})
}