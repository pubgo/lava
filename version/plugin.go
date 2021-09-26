package version

import (
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: "version",
		OnVars: func(w func(name string, data func() interface{})) {
			w("version", func() interface{} { return GetVer() })
		},
	})
}
