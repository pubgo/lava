package yuque

import (
	"testing"

	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/xerror"
)

var username = "barry.bai"
var userId = "253323"

func TestMain(t *testing.M) {
	defer xerror.RespExit()
	env.Load("../../../.env")
	t.Run()
}
