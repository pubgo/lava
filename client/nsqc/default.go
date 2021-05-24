package nsqc

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
)

var clients typex.SMap

func Get(names ...string) *nsqClient {
	var name = consts.GetDefault(names...)
	val, ok := clients.Load(name)
	if !ok {
		return nil
	}

	return val.(*nsqClient)
}

func initClient(name string, g Cfg) error {
	return try.Try(func() {
		xerror.Assert(name == "", "[name] should not be null")
		xerror.Assert(clients.Has(name), "nsq %s already exists", name)

		clients.Set(name, &nsqClient{cfg: g})
	})
}
