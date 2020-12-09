module github.com/pubgo/golug

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

require (
	github.com/HdrHistogram/hdrhistogram-go v1.0.0 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.5+incompatible
	github.com/andybalholm/brotli v1.0.1 // indirect
	github.com/apache/thrift v0.13.0
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/fasthttp/websocket v1.4.3
	github.com/fatedier/frp v0.34.3
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.4-0.20190803094908-fe8645b0a904+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fullstorydev/grpcurl v1.7.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/go-sql-driver/mysql v1.5.1-0.20200311113236-681ffa848bae
	github.com/gofiber/fiber/v2 v2.2.3
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/go-version v1.2.1
	github.com/imdario/mergo v0.3.11
	github.com/jhump/protoreflect v1.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/compress v1.11.3 // indirect
	github.com/lucas-clemente/quic-go v0.13.1
	github.com/mattn/go-sqlite3 v1.14.5
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/pubgo/dix v0.1.3
	github.com/pubgo/tikdog v0.0.0-20201130142326-26dbaa1f432c
	github.com/pubgo/xerror v0.3.1
	github.com/pubgo/xlog v0.0.10
	github.com/pubgo/xprocess v0.0.8
	github.com/pubgo/xprotogen v0.0.4
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rpcxio/libkv v0.4.2
	github.com/segmentio/nsq-go v1.2.4
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.6.1
	github.com/twmb/murmur3 v1.1.5
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	github.com/valyala/fasthttp v1.17.0
	github.com/valyala/fasttemplate v1.0.1
	github.com/vmihailenco/msgpack/v5 v5.0.0-rc.2
	go.etcd.io/etcd v0.0.0-20201125193152-8a03d2e9614b
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	google.golang.org/genproto v0.0.0-20201204160425-06b3db808446
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	vitess.io/vitess v0.7.0
	xorm.io/xorm v1.0.5
)
