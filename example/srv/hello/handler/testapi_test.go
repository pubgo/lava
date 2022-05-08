package handler

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/fx"

	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	"github.com/pubgo/lava/example/protopb/hellopb"
	"github.com/pubgo/lava/inject"
)

var _srv = &testApiHandler{}

func TestMain(t *testing.M) {
	inject.Register(fx.Populate(_srv))
	inject.Load()

	_srv.Init()
	t.Run()
}

func TestInit(t *testing.T) {
	fmt.Println(_srv.Version(context.Background(), &hellopb.TestReq{
		Input: "hello",
	}))
}
