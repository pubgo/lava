package orm

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	dix.Register(func(c config.Config) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = make(map[string]*Cfg)
		xerror.Panic(c.Decode(Name, &cfgMap))
		for name, cfg := range cfgMap {
			xerror.Panic(cfg.Valid())
			clients[name] = &Client{DB: cfg.Create()}
		}
		return clients
	})

	dix.Register(func(clients map[string]*Client) {
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
