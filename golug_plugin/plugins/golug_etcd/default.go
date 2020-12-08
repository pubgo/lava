package golug_etcd

import (
	"github.com/imdario/mergo"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

func GetClient(names ...string) (*clientv3.Client, error) {
	var name = golug_consts.Default
	if len(names) > 0 {
		name = names[0]
	}
	val, ok := clientM.Load(name)
	if !ok {
		return nil, xerror.Fmt("%s not found", name)
	}

	return val.(*clientv3.Client), nil
}

func initClient(name string, cfg ClientCfg) {
	_cfg := clientv3.Config{
		DialTimeout: Timeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	xerror.Panic(mergo.Map(&_cfg, cfg, mergo.WithOverride, mergo.WithAppendSlice))
	clientM.Store(name, xerror.PanicErr(clientv3.New(_cfg)).(*clientv3.Client))
}
