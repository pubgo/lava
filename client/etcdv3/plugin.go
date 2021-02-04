package etcdv3

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
)

func init() {
	golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
		},
		OnWatch: func(r *golug_watcher.Response) {

		},
	})
}

func NewWatcher() *etcdv3Watcher { return &etcdv3Watcher{} }

type etcdv3Watcher struct{}

func (e *etcdv3Watcher) GetPrefix() string { return Name }
func (e *etcdv3Watcher) OnDelete(key []byte) error {
	keys := strings.Split(string(key), ".")
	name := keys[len(keys)-1]

	log.Debugf("[etcd] delete client %s", name)
	if GetClient(name) == nil {
		log.Errorf("[etcd] client %s not found", name)
	}

	DelClient(name)
	return nil
}

func (e *etcdv3Watcher) OnPut(key []byte, value []byte) error {
	keys := strings.Split(string(key), ".")
	name := keys[len(keys)-1]

	log.Debugf("[etcd] update client %s", name)

	// 解析etcd配置
	var cfg config
	if err := json.Unmarshal(value, &cfg); err != nil {
		return errors.Wrapf(err, "[etcd] clientv3 Config parse error, cfg: %s", value)
	}

	return errors.Wrapf(UpdateClient(name, cfg.EtcdConfig()), "[etcd] client %s watcher update error", name)
}
