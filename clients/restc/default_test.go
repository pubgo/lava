package restc

import (
	"io/ioutil"
	"net/http"
	"testing"

	_ "github.com/pubgo/lava/encoding/json"
	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	defer xerror.RespTest(t)
	var resp, err = Get(nil, "http://baidu.com", func(req *Request) {})
	xerror.Panic(err)
	xerror.Assert(resp.Response().StatusCode != http.StatusOK, "code error")
	dt, err := ioutil.ReadAll(resp.Response().Body)
	xerror.Panic(err)
	t.Log(string(dt))
}
