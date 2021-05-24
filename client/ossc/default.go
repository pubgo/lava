package ossc

import (
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/env"
	"github.com/pubgo/xerror"
)

var clientM sync.Map

func GetClient(names ...string) *oss.Bucket {
	val, ok := clientM.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(*oss.Bucket)
}

func initClient(name string, cfg ClientCfg) {
	client, err := oss.New(
		env.Expand(cfg.Endpoint),
		env.Expand(cfg.AccessKeyID),
		env.Expand(cfg.AccessKeySecret),
	)
	xerror.Panic(err)
	kk := xerror.PanicErr(client.Bucket(cfg.Bucket)).(*oss.Bucket)
	clientM.Store(name, kk)
}
