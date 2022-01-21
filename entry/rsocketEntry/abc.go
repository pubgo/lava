package rsocketEntry

import (
	"github.com/pubgo/lava/entry"
	"google.golang.org/grpc"
)

type Entry interface {
	entry.Entry
	grpc.ServiceRegistrar
}
