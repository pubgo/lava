package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/pubgo/dix"
	"go.uber.org/fx"

	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	"github.com/pubgo/lava/example/protopb/hellopb"
)

var _srv = &testApiHandler{}

func TestMain(t *testing.M) {
	dix.Register(fx.Populate(_srv))
	dix.Invoke()

	_srv.Init()
	t.Run()
}

func TestInit(t *testing.T) {
	fmt.Println(_srv.Version(context.Background(), &hellopb.TestReq{
		Input: "hello",
	}))
}
