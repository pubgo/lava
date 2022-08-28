package hellorpc

import (
	"github.com/pubgo/lava/service"
)

func New() service.GrpcHandler {
	return &testApiHandler{}
}
