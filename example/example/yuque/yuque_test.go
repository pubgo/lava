package yuque

import (
	"testing"

	"github.com/pubgo/lava/example/protopb/proto/yuque_pb"
	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	var yq = yuque_pb.NewYuqueResty(cli)
	var _, err = yq.UserInfoByLogin(nil, &yuque_pb.UserInfoReq{Login: username})
	xerror.Panic(err)
}
