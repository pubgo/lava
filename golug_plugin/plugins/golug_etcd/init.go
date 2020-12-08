package golug_etcd

import (
	"sync"
	"time"
)

var name = "etcd"
var cfg Cfg
var clientM sync.Map

const Timeout = time.Second * 2
