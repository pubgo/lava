module github.com/pubgo/golug

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.3+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fullstorydev/grpcurl v1.7.0
	github.com/gofiber/fiber/v2 v2.2.3
	github.com/gofiber/template v1.6.6
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.5.3 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/hashicorp/go-version v1.2.1
	github.com/imdario/mergo v0.3.11
	github.com/jhump/protoreflect v1.7.1
	github.com/json-iterator/go v1.1.10
	github.com/klauspost/crc32 v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.9.10 // indirect
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/miekg/dns v1.1.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/pubgo/dix v0.1.13
	github.com/pubgo/xerror v0.3.25
	github.com/pubgo/xlog v0.0.16
	github.com/pubgo/xprocess v0.1.9
	github.com/pubgo/xprotogen v0.0.5
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/valyala/fasthttp v1.18.0 // indirect
	github.com/valyala/fasttemplate v1.0.1
	go.uber.org/atomic v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201203163018-be400aefbc4c
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/sys v0.0.0-20201204225414-ed752295db88 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/genproto v0.0.0-20201204160425-06b3db808446
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
	xorm.io/xorm v1.0.5
)
