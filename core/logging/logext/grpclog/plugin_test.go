package grpclog

import (
	"testing"

	"github.com/pubgo/funk/log"
	"google.golang.org/grpc/grpclog"
)

func TestName(t *testing.T) {
	New(log.GetLogger("test"))

	grpclog.Info("hello")
	grpclog.Component("cccc").Info("hello")
}
