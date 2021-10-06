module github.com/pubgo/lug/example/gid

go 1.16

replace github.com/pubgo/lug v0.1.5 => ../../

require (
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/mattheath/base62 v0.0.0-20150408093626-b80cdc656a7a // indirect
	github.com/mattheath/kala v0.0.0-20171219141654-d6276794bf0e
	github.com/pubgo/lug v0.1.5
	github.com/pubgo/xerror v0.4.11
	github.com/teris-io/shortid v0.0.0-20201117134242-e59966efd125
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)