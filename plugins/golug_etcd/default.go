package golug_etcd

import (
	"sync"

	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

var clientM sync.Map

func GetClient(names ...string) *clientv3.Client {
	var name = golug_consts.Default
	if len(names) > 0 && names[0] != "" {
		name = names[0]
	}
	val, ok := clientM.Load(name)
	if !ok {
		xerror.Next().Panic(xerror.Fmt("%s not found", name))
	}

	return val.(*clientv3.Client)
}

func initClient(name string, cfg ClientCfg) {
	_cfg := clientv3.Config{
		DialTimeout: Timeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	_cfg.Endpoints = cfg.Endpoints
	_cfg.AutoSyncInterval = cfg.AutoSyncInterval
	_cfg.DialTimeout = cfg.DialTimeout
	_cfg.DialKeepAliveTime = cfg.DialKeepAliveTime
	_cfg.DialKeepAliveTimeout = cfg.DialKeepAliveTimeout
	_cfg.Username = cfg.Username
	_cfg.Password = cfg.Password
	_cfg.RejectOldCluster = cfg.RejectOldCluster
	_cfg.PermitWithoutStream = cfg.PermitWithoutStream
	c := xerror.PanicErr(clientv3.New(_cfg)).(*clientv3.Client)
	clientM.Store(name, c)
}
