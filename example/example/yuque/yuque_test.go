package yuque

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	yuque_pb "github.com/pubgo/lava/example/protopb/proto/yuque"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	var cli = resty.New().
		SetDebug(runenv.IsDev()).
		SetContentLength(true).
		SetBaseURL("https://www.yuque.com/api/v2").
		SetRetryCount(3).
		SetTimeout(time.Second * 2).
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
			"User-Agent":   "test",
			"X-Auth-Token": env.Get("yuque_token"),
		})

	var yq = yuque_pb.NewYuqueResty(cli)
	var resp, err = yq.UserInfoByLogin(nil, &yuque_pb.UserInfoReq{Login: "barry.bai"})
	xerror.Panic(err)
	fmt.Println(resp)
}
