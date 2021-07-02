package plugin

import (
	"github.com/pubgo/lug/vars"

	"net/url"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var data = url.Values{}
		for k, v := range All() {
			for i := range v {
				data.Add(k, v[i].String())
			}
		}
		return data
	})
}
