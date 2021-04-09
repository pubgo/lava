package etcd

import (
	"time"

	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcd"

var cfgList = make(map[string]Cfg)

type Cfg struct {
	Endpoints            []string          `json:"endpoints"`
	AutoSyncInterval     time.Duration     `json:"interval"`
	DialTimeout          time.Duration     `json:"timeout"`
	DialKeepAliveTime    time.Duration     `json:"keepalive"`
	DialKeepAliveTimeout time.Duration     `json:"keepalive_timeout"`
	MaxCallSendMsgSize   int               `json:"max_send"`
	MaxCallRecvMsgSize   int               `json:"max_recv"`
	Username             string            `json:"username"`
	Name                 string            `json:"name"`
	Password             string            `json:"password"`
	DialOptions          []grpc.DialOption `json:"-"`
}

// 转化为etcd Cfg
func (t Cfg) ToEtcd() (cfg clientv3.Config) {
	xerror.Panic(merge.Copy(&cfg, &t))
	return
}

func GetDefaultCfg() Cfg {
	return Cfg{
		DialTimeout: time.Second * 2,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
}
