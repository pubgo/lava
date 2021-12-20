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
	var resp, err = Get(nil, "http://baidu.com")
	xerror.Panic(err)
	xerror.Assert(resp.StatusCode != http.StatusOK, "code error")
	dt, err := ioutil.ReadAll(resp.Body)
	xerror.Panic(err)
	t.Log(string(dt))
}
