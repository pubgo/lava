package ossc

import (
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/resource"
)

func Get(names ...string) *Client {
	val := resource.Get(Name, lavax.GetDefault(names...))
	if val == nil {
		return nil
	}

	return val.(*Client)
}
