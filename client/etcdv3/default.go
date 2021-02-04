package etcdv3

import (
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
)

const Name = "etcdv3"
const etcdEnvPrefix = "GAPI_ETCD_"
const DefaultTimeout = time.Second * 2

var data sync.Map
var DefaultCfg = clientv3.Config{
	DialTimeout: DefaultTimeout,
	DialOptions: []grpc.DialOption{grpc.WithBlock()},
}
