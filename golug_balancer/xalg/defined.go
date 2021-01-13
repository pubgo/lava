package xalg

import "google.golang.org/grpc/balancer"

//all xalg interface

type P2c interface {
	Next() (interface{}, func(balancer.DoneInfo))
	Add(item interface{})
}

type WeightRound interface {
	Next() (interface{}, func(balancer.DoneInfo))
	Add(name string, weight int64, item interface{})
}
