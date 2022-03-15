package rsocketEntry

import (
	"github.com/pubgo/lava/server"
	"google.golang.org/grpc"
)

type Entry interface {
	server.Entry
	grpc.ServiceRegistrar
}
