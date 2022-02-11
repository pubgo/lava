package ossc

import (
	"github.com/pubgo/lava/pkg/utils"
	"github.com/pubgo/lava/resource"
)

func Get(names ...string) *Client {
	val := resource.Get(Name, utils.GetDefault(names...))
	if val == nil {
		return nil
	}

	return val.(*Client)
}
