package etcdv3

import (
	"time"

	"github.com/pubgo/x/jsonx"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"

var cfgList []Cfg

type Cfg struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     jsonx.Duration    `json:"interval"`
	DialTimeout          jsonx.Duration    `json:"timeout"`
	DialKeepAliveTime    jsonx.Duration    `json:"keepalive"`
	DialKeepAliveTimeout jsonx.Duration    `json:"keepalive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send"`
	MaxCallRecvMsgSize   int               `json:"max_recv"`
	Username             string            `json:"username"`
	Name                 string            `json:"name"`
	Password             string            `json:"password"`
	DialOptions          []grpc.DialOption `json:"-"`
}

// 转化为etcd Cfg
func (t Cfg) ToEtcdConfig() (cfg clientv3.Config) {
	cfg.Endpoints = t.Endpoints
	cfg.AutoSyncInterval = t.AutoSyncInterval.Duration
	cfg.DialTimeout = t.DialTimeout.Duration
	cfg.DialKeepAliveTime = t.DialKeepAliveTime.Duration
	cfg.DialKeepAliveTimeout = t.DialKeepAliveTimeout.Duration
	cfg.MaxCallSendMsgSize = t.MaxCallSendMsgSize
	cfg.MaxCallRecvMsgSize = t.MaxCallRecvMsgSize
	cfg.Username = t.Username
	cfg.Password = t.Password
	cfg.DialOptions = t.DialOptions
	return cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{
		DialTimeout: jsonx.Dur(time.Second * 2),
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
