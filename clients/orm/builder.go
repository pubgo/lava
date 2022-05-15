package orm

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
	"github.com/pubgo/xerror"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, &cfgMap))
	for name := range cfgMap {
		cfg := cfgMap[name]
		xerror.Panic(cfg.Valid())
		inject.NameGroup(Name, name, func() *Client {
			return &Client{DB: cfg.Get()}
		})
	}
}

func init() {
	inject.Invoke(func(clients ...*Client) {
		vars.Register(Name+"_stats", func() interface{} {
			var data typex.A
			for i := range clients {
				var ss = make(typex.M)
				cli := clients[i]
				_db, err := cli.DB.DB()
				if err != nil {
					ss["data"] = err.Error()
				} else {
					ss["data"] = _db.Stats()
				}
				data.Append(ss)
			}
			return data
		})
	})
}
