package golug_etcd

import (
	"context"
	"github.com/pubgo/golug/golug_consts"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

func GetCfg() Cfg {
	return cfg
}

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
	xerror.PanicErr(c.AuthStatus(context.Background()))
}
