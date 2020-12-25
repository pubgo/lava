module github.com/pubgo/golug

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.0 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.5+incompatible
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/fasthttp/websocket v1.4.3-beta.4
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.3+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fullstorydev/grpcurl v1.7.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/gofiber/fiber/v2 v2.2.3
	github.com/gofiber/template v1.6.6
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/go-version v1.2.1
	github.com/imdario/mergo v0.3.11
	github.com/jhump/protoreflect v1.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/crc32 v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.9.9 // indirect
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/mmcloughlin/avo v0.0.0-20201216231306-039ef47f4f69 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pubgo/dix v0.1.8
	github.com/pubgo/xerror v0.3.4
	github.com/pubgo/xlog v0.0.10
	github.com/pubgo/xprocess v0.0.11
	github.com/pubgo/xprotogen v0.0.4
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/rpcxio/libkv v0.4.2
	github.com/segmentio/nsq-go v1.2.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	github.com/valyala/fasthttp v1.18.0
	github.com/valyala/fasttemplate v1.0.1
	go.etcd.io/etcd v0.0.0-20201125193152-8a03d2e9614b
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	google.golang.org/genproto v0.0.0-20201204160425-06b3db808446
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	vitess.io/vitess v0.7.0
	xorm.io/xorm v1.0.5
)
