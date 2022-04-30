package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"

	_ "github.com/pubgo/lava/clients/orm/driver/sqlite"

	"github.com/pubgo/lava/example/protopb/proto/hello"
)

var _srv = &testApiHandler{}

func TestMain(t *testing.M) {
	inject.Init(append(inject.List(), fx.Populate(_srv))...)

	_srv.Init()
	t.Run()
}

func TestInit(t *testing.T) {
	fmt.Println(_srv.Version(context.Background(), &hello.TestReq{
		Input: "hello",
	}))
}
