module github.com/pubgo/lug

go 1.16

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.8+incompatible
	github.com/antonmedv/expr v1.9.0
	github.com/arl/statsviz v0.4.0
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/emicklei/proto v1.9.1
	github.com/fasthttp/websocket v1.4.3-rc.3
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.3+incompatible
	github.com/fatih/color v1.9.0
	github.com/favadi/protoc-go-inject-tag v1.3.0
	github.com/felixge/fgprof v0.9.1
	github.com/flosch/pongo2/v4 v4.0.2
	github.com/fogleman/gg v1.2.1-0.20190220221249-0403632d5b90
	github.com/fullstorydev/grpcurl v1.8.2
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-chi/chi/v5 v5.0.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/go-echarts/statsview v0.3.4
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/go-openapi/loads v0.19.5
	github.com/go-openapi/runtime v0.19.21
	github.com/go-openapi/spec v0.20.2
	github.com/go-openapi/swag v0.19.13 // indirect
	github.com/go-redis/redis/v8 v8.8.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gofiber/fiber/v2 v2.12.0
	github.com/gofiber/template v1.6.6
	github.com/gofiber/websocket/v2 v2.0.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/google/gops v0.3.18
	github.com/google/uuid v1.2.0
	github.com/gordonklaus/ineffassign v0.0.0-20200309095847-7953dde2c7bf
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/schema v1.2.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/memberlist v0.1.3
	github.com/iancoleman/strcase v0.2.0
	github.com/jaegertracing/jaeger v1.22.0
	github.com/jhump/protoreflect v1.9.0
	github.com/json-iterator/go v1.1.11
	github.com/klauspost/crc32 v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.9.10 // indirect
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/magefile/mage v1.11.0
	github.com/maragudk/gomponents v0.17.2
	github.com/maruel/panicparse/v2 v2.1.1
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/mattn/go-zglob v0.0.3
	github.com/mitchellh/hashstructure v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/nacos-group/nacos-sdk-go v1.0.9
	github.com/olekukonko/tablewriter v0.0.2-0.20190409134802-7e037d187b0c
	github.com/opentracing/opentracing-go v1.1.0
	github.com/panjf2000/gnet v1.4.5
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/pubgo/dix v0.1.28
	github.com/pubgo/x v0.3.36
	github.com/pubgo/xerror v0.4.11
	github.com/pubgo/xlog v0.2.8
	github.com/pubgo/xprotogen v0.0.17 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/savsgio/gotils v0.0.0-20210520110740-c57c45b83e0a // indirect
	github.com/segmentio/ksuid v1.0.3
	github.com/segmentio/nsq-go v1.2.4
	github.com/shirou/gopsutil/v3 v3.21.2
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/swaggo/http-swagger v1.1.1 // indirect
	github.com/tinylib/msgp v1.1.6
	github.com/toqueteos/webbrowser v1.2.0 // indirect
	github.com/twmb/murmur3 v1.1.5 // indirect
	github.com/uber-go/tally v3.4.2+incompatible
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	github.com/valyala/fasthttp v1.26.0
	github.com/valyala/fasttemplate v1.2.1
	github.com/vmihailenco/msgpack/v5 v5.3.1
	go.etcd.io/bbolt v1.3.3
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/atomic v1.7.0
	go.uber.org/automaxprocs v1.4.0
	go.uber.org/zap v1.19.0
	golang.org/x/image v0.0.0-20191206065243-da761ea9ff43
	golang.org/x/net v0.0.0-20210510120150-4163338589ed
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40
	golang.org/x/tools v0.1.3
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.38.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/grpc/examples v0.0.0-20210622215705-4440c3b8306d // indirect
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/klog/v2 v2.9.0 // indirect
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.1 // indirect
	xorm.io/xorm v1.0.5
)
