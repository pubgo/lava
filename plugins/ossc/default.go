package ossc

import (
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/consts"
)

var clientM sync.Map

func Get(names ...string) *oss.Bucket {
	val, ok := clientM.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(*oss.Bucket)
}

func Update(name string, cfg ClientCfg) (err error) {
	defer xerror.RespErr(&err)
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	xerror.Panic(err)
	kk := xerror.PanicErr(client.Bucket(cfg.Bucket)).(*oss.Bucket)
	clientM.Store(name, kk)
	return nil
}
