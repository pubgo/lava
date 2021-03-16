module github.com/pubgo/golug

go 1.15

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

replace github.com/bufbuild/buf => github.com/bufbuild/buf v0.36.0

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/bufbuild/buf v0.37.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.3+incompatible
	github.com/go-chi/chi/v5 v5.0.0
	github.com/gofiber/fiber/v2 v2.2.3
	github.com/gofiber/template v1.6.6
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.3-0.20201103224600-674baa8c7fc3
	github.com/golangci/golangci-lint v1.28.3
	github.com/google/uuid v1.2.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway v1.14.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.3.0
	github.com/hashicorp/go-version v1.2.1
	github.com/imdario/mergo v0.3.11
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/crc32 v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.9.10 // indirect
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/prometheus/client_golang v1.4.1
	github.com/pubgo/dix v0.1.14
	github.com/pubgo/x v0.3.18-0.20210308075649-51ffee8d36c4
	github.com/pubgo/xerror v0.4.1
	github.com/pubgo/xlog v0.0.21
	github.com/pubgo/xprotogen v0.0.7
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/valyala/fasthttp v1.18.0 // indirect
	github.com/valyala/fasttemplate v1.0.1
	go.etcd.io/etcd v0.0.0-20200402134248-51bdeb39e698
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/genproto v0.0.0-20210224155714-063164c882e6
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.1-0.20201208041424-160c7477e0e8
	xorm.io/xorm v1.0.5
)
