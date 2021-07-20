package ossc

import (
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/consts"
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
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	xerror.Panic(err)
	kk := xerror.PanicErr(client.Bucket(cfg.Bucket)).(*oss.Bucket)
	clientM.Store(name, kk)
}
