module github.com/pubgo/lava

go 1.17

replace github.com/go-echarts/statsview => github.com/pubgo/statsview v0.3.5

require (
	filippo.io/mkcert v1.4.3
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/aliyun/aliyun-oss-go-sdk v2.1.8+incompatible
	github.com/antonmedv/expr v1.9.0
	github.com/arl/statsviz v0.4.0
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/bigwhite/functrace v0.0.0-20210622013229-318a19dbb29a
	github.com/bojand/ghz v0.105.0
	github.com/casbin/casbin/v2 v2.1.2
	github.com/charmbracelet/glow v1.4.1
	github.com/edsrzf/mmap-go v1.0.0
	github.com/emicklei/proto v1.9.1
	github.com/fasthttp/websocket v1.4.3-rc.3
	github.com/fatedier/golib v0.2.0
	github.com/fatedier/kcp-go v2.0.3+incompatible
	github.com/fatih/color v1.12.0
	github.com/favadi/protoc-go-inject-tag v1.3.0
	github.com/felixge/fgprof v0.9.1
	github.com/flosch/pongo2/v4 v4.0.2
	github.com/fogleman/gg v1.2.1-0.20190220221249-0403632d5b90
	github.com/fullstorydev/grpcui v1.2.0
	github.com/fullstorydev/grpcurl v1.8.5
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-chi/chi/v5 v5.0.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/go-echarts/statsview v0.3.4
	github.com/go-openapi/loads v0.19.5
	github.com/go-openapi/runtime v0.19.21
	github.com/go-openapi/spec v0.20.2
	github.com/go-redis/redis/v8 v8.8.3
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gofiber/fiber/v2 v2.19.0
	github.com/gofiber/template v1.6.6
	github.com/gofiber/websocket/v2 v2.0.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gops v0.3.18
	github.com/google/uuid v1.3.0
	github.com/gordonklaus/ineffassign v0.0.0-20210225214923-2e10b2664254
	github.com/gorilla/schema v1.2.0
	github.com/grandcat/zeroconf v1.0.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/memberlist v0.1.3
	github.com/iancoleman/strcase v0.2.0
	github.com/jaegertracing/jaeger v1.22.0
	github.com/jhump/protoreflect v1.10.1
	github.com/jinzhu/copier v0.2.8
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5
	github.com/json-iterator/go v1.1.11
	github.com/klauspost/crc32 v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.9.10 // indirect
	github.com/lib/pq v1.10.1
	github.com/lucas-clemente/quic-go v0.19.3
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/maragudk/gomponents v0.17.2
	github.com/maruel/panicparse/v2 v2.1.1
	github.com/mattheath/kala v0.0.0-20171219141654-d6276794bf0e
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/mattn/go-zglob v0.0.3
	github.com/mitchellh/hashstructure v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/nacos-group/nacos-sdk-go v1.0.9
	github.com/olekukonko/tablewriter v0.0.5
	github.com/opentracing/opentracing-go v1.1.0
	github.com/panjf2000/gnet v1.4.5
	github.com/pelletier/go-toml v1.9.3
	github.com/piotrkowalczuk/promgrpc/v4 v4.0.4
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/pubgo/dix v0.1.31
	github.com/pubgo/x v0.3.36
	github.com/pubgo/xerror v0.4.14
	github.com/pubgo/xlog v0.2.10
	github.com/reugn/go-quartz v0.3.7
	github.com/savsgio/gotils v0.0.0-20210520110740-c57c45b83e0a // indirect
	github.com/segmentio/ksuid v1.0.3
	github.com/segmentio/nsq-go v1.2.4
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/http-swagger v1.1.1
	github.com/teris-io/shortid v0.0.0-20201117134242-e59966efd125
	github.com/tidwall/btree v0.6.1
	github.com/tinylib/msgp v1.1.6
	github.com/twmb/murmur3 v1.1.5 // indirect
	github.com/uber-go/tally v3.4.2+incompatible
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	github.com/valyala/fasthttp v1.29.0
	github.com/valyala/fastrand v1.1.0
	github.com/vmihailenco/msgpack/v5 v5.3.1
	github.com/webview/webview v0.0.0-20210330151455-f540d88dde4e
	go.etcd.io/bbolt v1.3.5
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/atomic v1.7.0
	go.uber.org/automaxprocs v1.4.0
	go.uber.org/zap v1.19.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/image v0.0.0-20191206065243-da761ea9ff43
	golang.org/x/mod v0.5.0
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210908233432-aa78b53d3365
	golang.org/x/tools v0.1.5
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.41.0
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

require (
	cloud.google.com/go v0.81.0 // indirect
	github.com/AlecAivazis/survey/v2 v2.3.2
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/alecthomas/chroma v0.8.2 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.18 // indirect
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/armon/go-metrics v0.0.0-20180917152333-f0300d1749da // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/atotto/clipboard v0.1.2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blevesearch/bleve/v2 v2.2.2
	github.com/buger/jsonparser v0.0.0-20181115193947-bf1c66bbce23 // indirect
	github.com/c-bata/go-prompt v0.2.6
	github.com/calmh/randomart v1.1.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/charmbracelet/bubbles v0.7.6 // indirect
	github.com/charmbracelet/bubbletea v0.13.2 // indirect
	github.com/charmbracelet/charm v0.8.6 // indirect
	github.com/charmbracelet/glamour v0.2.1-0.20210402234443-abe9cda419ba // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/chris-ramon/douceur v0.2.0 // indirect
	github.com/chriswalz/bit v1.1.2
	github.com/chyroc/go-aliyundrive v0.1.0
	github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403 // indirect
	github.com/cncf/xds/go v0.0.0-20210805033703-aa0b78936158 // indirect
	github.com/containerd/console v1.0.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/dop251/goja v0.0.0-20200721192441-a695b0cdd498
	github.com/dustin/go-humanize v1.0.1-0.20200219035652-afde56e7acac // indirect
	github.com/envoyproxy/go-control-plane v0.9.10-0.20210907150352-cf90f659a021 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/evanw/esbuild v0.13.12
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-openapi/analysis v0.19.10 // indirect
	github.com/go-openapi/errors v0.19.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.13 // indirect
	github.com/go-openapi/validate v0.19.10 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.1.1-0.20200604201612-c04b05f3adfa // indirect
	github.com/google/pprof v0.0.0-20210226084205-cbba55b83ad5 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/igm/sockjs-go/v3 v3.0.1
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/improbable-eng/grpc-web v0.12.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/josharian/intern v1.0.0 // indirect
	github.com/keybase/go-ps v0.0.0-20190827175125-91aafc93ba19 // indirect
	github.com/kisielk/errcheck v1.6.0 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/klauspost/cpuid/v2 v2.0.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/ktr0731/evans v0.10.0
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f // indirect
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/magefile/mage v1.11.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/manifoldco/promptui v0.8.0
	github.com/marten-seemann/qtls v0.10.0 // indirect
	github.com/marten-seemann/qtls-go1-15 v0.1.1 // indirect
	github.com/mattheath/base62 v0.0.0-20150408093626-b80cdc656a7a // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/maxence-charriere/go-app/v9 v9.1.2
	github.com/meowgorithm/babyenv v1.3.1 // indirect
	github.com/microcosm-cc/bluemonday v1.0.5 // indirect
	github.com/miekg/dns v1.1.35 // indirect
	github.com/mikesmitty/edkey v0.0.0-20170222072505-3356ea4e686a // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-ps v1.0.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/muesli/gitcha v0.2.0 // indirect
	github.com/muesli/go-app-paths v0.2.1 // indirect
	github.com/muesli/reflow v0.2.1-0.20210115123740-9e1d0d53df68 // indirect
	github.com/muesli/sasquatch v0.0.0-20200811221207-66979d92330a // indirect
	github.com/muesli/termenv v0.8.1 // indirect
	github.com/nxadm/tail v1.4.8
	github.com/open2b/scriggo v0.53.4
	github.com/panjf2000/ants/v2 v2.4.5 // indirect
	github.com/philhofer/fwd v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/quickjs-go/quickjs-go v0.0.0-20210519010351-d203c0c75f0a
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rs/cors v1.7.0
	github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94 // indirect
	github.com/sahilm/fuzzy v0.1.0 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/swaggo/files v0.0.0-20190704085106-630677cd5c14 // indirect
	github.com/swaggo/swag v1.7.0 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20200815110645-5c35d600f0ca // indirect
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/tklauser/numcpus v0.2.1 // indirect
	github.com/toolkits/concurrent v0.0.0-20150624120057-a4371d70e3e3 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xlab/treeprint v1.0.0 // indirect
	github.com/yuin/goldmark v1.4.1 // indirect
	github.com/yuin/goldmark-emoji v1.0.1 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.0 // indirect
	go.mongodb.org/mongo-driver v1.3.4 // indirect
	go.opentelemetry.io/otel v0.20.0 // indirect
	go.opentelemetry.io/otel/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/trace v0.20.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/oauth2 v0.0.0-20210615190721-d04028783cf1 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	honnef.co/go/tools v0.2.0 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
	k8s.io/api v0.21.1 // indirect
	rsc.io/goversion v1.2.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
	software.sslmate.com/src/go-pkcs12 v0.0.0-20180114231543-2291e8f0f237 // indirect
	xorm.io/builder v0.3.7 // indirect
)

require (
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/apex/log v1.9.0 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/blevesearch/bleve_index_api v1.0.1 // indirect
	github.com/blevesearch/go-porterstemmer v1.0.3 // indirect
	github.com/blevesearch/mmap-go v1.0.3 // indirect
	github.com/blevesearch/scorch_segment_api/v2 v2.1.0 // indirect
	github.com/blevesearch/segment v0.9.0 // indirect
	github.com/blevesearch/snowballstem v0.9.0 // indirect
	github.com/blevesearch/upsidedown_store_api v1.0.1 // indirect
	github.com/blevesearch/vellum v1.0.7 // indirect
	github.com/blevesearch/zapx/v11 v11.3.1 // indirect
	github.com/blevesearch/zapx/v12 v12.3.1 // indirect
	github.com/blevesearch/zapx/v13 v13.3.1 // indirect
	github.com/blevesearch/zapx/v14 v14.3.1 // indirect
	github.com/blevesearch/zapx/v15 v15.3.1 // indirect
	github.com/c4milo/unpackit v0.0.0-20170704181138-4ed373e9ef1c // indirect
	github.com/chriswalz/complete/v3 v3.0.13 // indirect
	github.com/chyroc/gorequests v0.26.0 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/couchbase/ghistogram v0.1.0 // indirect
	github.com/couchbase/moss v0.1.0 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/go-sourcemap/sourcemap v2.1.2+incompatible // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/juju/ansiterm v0.0.0-20180109212912-720a0952cc2a // indirect
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/juju/persistent-cookiejar v0.0.0-20171026135701-d5e5a8405ef9 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/ktr0731/go-multierror v0.0.0-20171204182908-b7773ae21874 // indirect
	github.com/ktr0731/go-prompt v0.2.2-0.20190609072126-7894cc3f2925 // indirect
	github.com/ktr0731/go-shellstring v0.1.3 // indirect
	github.com/ktr0731/go-updater v0.1.5 // indirect
	github.com/ktr0731/grpc-web-go-client v0.2.7 // indirect
	github.com/lithammer/fuzzysearch v1.1.1 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-pipeline v0.0.0-20190323144519-32d779b32768 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/mwitkow/grpc-proxy v0.0.0-20181017164139-0f1106ef9c76 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/posener/script v1.1.5 // indirect
	github.com/rogpeppe/go-internal v1.6.2 // indirect
	github.com/rs/zerolog v1.20.0 // indirect
	github.com/shirou/gopsutil/v3 v3.21.5 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/steveyen/gtreap v0.1.0 // indirect
	github.com/thoas/go-funk v0.7.0 // indirect
	github.com/tj/go-spin v1.1.0 // indirect
	github.com/tj/go-update v2.2.4+incompatible // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/zchee/go-xdgbasedir v1.0.3 // indirect
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)
