package orm

import (
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/vars"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		BuilderFactory: resource.Factory{
			DefaultCfg: DefaultCfg(),
			ResType:    &Client{},
		},
		OnVars: func(v vars.Publisher) {
			v.Publish(Name+"_stats", func() interface{} {
				var data = make(map[string]interface{})
				for k, v := range resource.GetByKind(Name) {
					db, err := v.(*Client).get().DB()
					if err != nil {
						data[k] = err.Error()
					} else {
						data[k] = db.Stats()
					}
				}
				return data
			})
		},
	})
}
