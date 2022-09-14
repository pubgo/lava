package gidsrv

import "github.com/pubgo/lava/example/gen/proto/gidpb"

type Service interface {
	gidpb.IdServer
}
