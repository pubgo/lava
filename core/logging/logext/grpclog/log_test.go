package grpclog

import (
	"testing"

	"google.golang.org/grpc/grpclog"

	"github.com/pubgo/funk/v2/log"
)

func TestName(t *testing.T) {
	SetLogger(log.GetLogger("test"))

	grpclog.Info("hello")
	grpclog.Component("cccc").Info("hello")
}
