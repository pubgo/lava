module github.com/pubgo/lava

go 1.19

replace github.com/google/gnostic => github.com/google/gnostic v0.7.0

require (
	github.com/gofiber/fiber/v2 v2.51.0
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grandcat/zeroconf v1.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/json-iterator/go v1.1.12
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/errors v0.9.1
	github.com/pubgo/opendoc v0.0.4-3
	github.com/reugn/go-quartz v0.3.7
	github.com/testcontainers/testcontainers-go v0.28.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.28.0
	github.com/vmihailenco/msgpack/v5 v5.3.1
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/metric v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.19.0
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20231212172506-995d672761c0 // indirect
	google.golang.org/grpc v1.61.0
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/goccy/go-json v0.10.2
	github.com/mailgun/holster/v4 v4.11.0
	github.com/prometheus/client_golang v1.18.0
	github.com/valyala/bytebufferpool v1.0.0
	gorm.io/gorm v1.24.5
)

require (
	ariga.io/atlas v0.10.0
	entgo.io/ent v0.12.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/arl/statsviz v0.6.0
	github.com/dave/jennifer v1.6.0
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/fasthttp/websocket v1.5.7
	github.com/felixge/fgprof v0.9.3
	github.com/fullstorydev/grpchan v1.1.1
	github.com/go-playground/validator/v10 v10.10.1
	github.com/gobwas/ws v1.3.2
	github.com/gofiber/adaptor/v2 v2.1.30
	github.com/gofiber/contrib/websocket v1.3.0
	github.com/gofiber/template v1.6.25
	github.com/gofiber/utils v1.0.1
	github.com/gofiber/websocket/v2 v2.1.5
	github.com/google/gnostic v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19
	github.com/libp2p/go-nat v0.2.0
	github.com/maragudk/gomponents v0.20.0
	github.com/mattheath/kala v0.0.0-20171219141654-d6276794bf0e
	github.com/pubgo/dix v0.3.15-0.20240107153647-472348fb7f95
	github.com/pubgo/funk v0.5.39-0.20240218021552-f8223b071505
	github.com/rs/xid v1.5.0
	github.com/rs/zerolog v1.30.0
	github.com/stretchr/testify v1.8.4
	github.com/teris-io/shortid v0.0.0-20220617161101-71ec9f2aa569
	github.com/tidwall/match v1.1.1
	github.com/uber-go/tally/v4 v4.1.7
	github.com/urfave/cli/v3 v3.0.0-alpha2
	github.com/valyala/fasthttp v1.51.0
	github.com/valyala/fasttemplate v1.2.2
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.23.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/exporters/prometheus v0.45.2
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.23.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.23.1
	go.opentelemetry.io/otel/sdk v1.23.1
	go.opentelemetry.io/otel/sdk/metric v1.23.1
	google.golang.org/genproto/googleapis/api v0.0.0-20240102182953-50ed04b92917
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240102182953-50ed04b92917
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.4.5
	gorm.io/driver/postgres v1.4.1
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gen v0.3.21
	nhooyr.io/websocket v1.8.6
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/Microsoft/hcsshim v0.11.4 // indirect
	github.com/a8m/envsubst v1.3.0 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/alecthomas/repr v0.2.0 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/containerd/containerd v1.7.12 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/distribution/reference v0.5.0 // indirect
	github.com/docker/docker v25.0.2+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/getkin/kin-openapi v0.115.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.22.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gnostic-models v0.6.9-0.20230804172637-c7be7c783f49 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/google/pprof v0.0.0-20230602150820-91b7bce49751 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl/v2 v2.13.0 // indirect
	github.com/huin/goupnp v1.2.0 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jhump/protoreflect v1.11.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/k0kubun/pp/v3 v3.2.0 // indirect
	github.com/klauspost/compress v1.17.3 // indirect
	github.com/koron/go-ssdp v0.0.4 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/libp2p/go-netroute v0.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattheath/base62 v0.0.0-20150408093626-b80cdc656a7a // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc5 // indirect
	github.com/perimeterx/marshmallow v1.1.4 // indirect
	github.com/phuslu/goid v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twmb/murmur3 v1.1.5 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	github.com/zclconf/go-cty v1.8.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0 // indirect
	go.opentelemetry.io/proto/otlp v1.1.0 // indirect
	golang.org/x/exp v0.0.0-20230811145659-89c5cff77bcb // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.10.0 // indirect
	gorm.io/datatypes v1.0.7 // indirect
	gorm.io/hints v1.1.0 // indirect
	gorm.io/plugin/dbresolver v1.3.0 // indirect
	k8s.io/kube-openapi v0.0.0-20221123214604-86e75ddd809a // indirect
)
