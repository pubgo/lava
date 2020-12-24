package grpclient

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/pkg/golug_utils"
)

func init() {
	golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnInit: func(ent golug_entry.Entry) {
			golug_config.Decode(Name, &cfg)

			for k, v := range cfg {
				_cfg := GetDefaultCfg()
				golug_utils.Mergo(&_cfg, v)
				cfg[k] = _cfg

				value, _ := connPool.LoadOrStore(k, &grpcPool{cfg: _cfg, addr: k})
				pool := value.(*grpcPool)
				// 服务启动之前, 初始化grpc conn pool
				for i := 5; i > 0; i-- {
					cc := pool.createConn()
					pool.connList = append(pool.connList, cc)
					pool.connMap.Store(cc, struct{}{})
				}
			}
		},
	})
}
