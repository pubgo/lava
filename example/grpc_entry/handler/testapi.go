package handler

import (
	"context"

	"github.com/pubgo/lug/example/proto/hello"
)

func NewTestAPIHandler() hello.TestApiServer {
	return &testapiHandler{}
}

type testapiHandler struct {
}

func (h *testapiHandler) Version(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {
	//log.Infof("Received Helloworld.Call request, name: %s", in.Input)

	out = &hello.TestApiOutput{
		Msg: in.Input,
	}
	return
}

func (h *testapiHandler) VersionTest(ctx context.Context, in *hello.TestReq) (out *hello.TestApiOutput, err error) {

	out = &hello.TestApiOutput{
		Msg: in.Input + "_test",
	}
	return
}
