package yuque

import (
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/xerror"
)

var username = "barry.bai"
var userId = "253323"
var cli *resty.Client

func TestMain(t *testing.M) {
	defer xerror.RespExit()
	env.Load("../../../.env")

	cli = resty.New().
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
	t.Run()
}
