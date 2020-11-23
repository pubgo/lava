package golug_tracing

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	jaegerProm "github.com/uber/jaeger-lib/metrics/prometheus"
)

const (
	// environment variable names
	envServiceName                         = "JAEGER_SERVICE_NAME"
	envDisabled                            = "JAEGER_DISABLED"
	envRPCMetrics                          = "JAEGER_RPC_METRICS"
	envTags                                = "JAEGER_TAGS"
	envSamplerType                         = "JAEGER_SAMPLER_TYPE"
	envSamplerParam                        = "JAEGER_SAMPLER_PARAM"
	envSamplerManagerHostPort              = "JAEGER_SAMPLER_MANAGER_HOST_PORT" // Deprecated by envSamplingEndpoint
	envSamplingEndpoint                    = "JAEGER_SAMPLING_ENDPOINT"
	envSamplerMaxOperations                = "JAEGER_SAMPLER_MAX_OPERATIONS"
	envSamplerRefreshInterval              = "JAEGER_SAMPLER_REFRESH_INTERVAL"
	envReporterMaxQueueSize                = "JAEGER_REPORTER_MAX_QUEUE_SIZE"
	envReporterFlushInterval               = "JAEGER_REPORTER_FLUSH_INTERVAL"
	envReporterLogSpans                    = "JAEGER_REPORTER_LOG_SPANS"
	envReporterAttemptReconnectingDisabled = "JAEGER_REPORTER_ATTEMPT_RECONNECTING_DISABLED"
	envReporterAttemptReconnectInterval    = "JAEGER_REPORTER_ATTEMPT_RECONNECT_INTERVAL"
	envEndpoint                            = "JAEGER_ENDPOINT"
	envUser                                = "JAEGER_USER"
	envPassword                            = "JAEGER_PASSWORD"
	envAgentHost                           = "JAEGER_AGENT_HOST"
	envAgentPort                           = "JAEGER_AGENT_PORT"
)

var name = "tracing"
var cfg = xerror.PanicErr(jaegerCfg.FromEnv()).(*jaegerCfg.Configuration)

func Middleware(ctx *fiber.Ctx) error {
	var headers = make(http.Header)
	ctx.Request().Header.VisitAll(func(key, value []byte) { headers.Add(string(key), string(value)) })

	operationName := ctx.Route().Method + ":" + ctx.Route().Path
	var parentSpan opentracing.Span
	spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(headers))
	if err != nil {
		parentSpan = opentracing.GlobalTracer().StartSpan(operationName)
	} else {
		parentSpan = opentracing.StartSpan(operationName,
			opentracing.ChildOf(spCtx),
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			ext.SpanKindRPCServer,
		)
	}

	_ctx := context.Background()
	ctx.Context().VisitUserValues(func(bytes []byte, i interface{}) {
		_ctx = context.WithValue(_ctx, string(bytes), i)
	})

	ctx.Context().SetUserValue("ParentSpanContext", opentracing.ContextWithSpan(_ctx, parentSpan))

	xerror.Panic(opentracing.GlobalTracer().Inject(parentSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(headers)))

	parentSpan.SetTag("http.method", ctx.Method())
	parentSpan.SetTag("http.url", ctx.OriginalURL())
	parentSpan.SetTag("http.request.host", ctx.Hostname())

	defer parentSpan.Finish()
	defer func() {
		parentSpan.SetTag(string(ext.HTTPStatusCode), ctx.Response().StatusCode())
	}()

	return xerror.Wrap(ctx.Next())
}

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.Decode(name, cfg))

			xerror.Panic(ent.UnWrap(func(entry golug_entry.HttpEntry) { entry.Use(Middleware) }))

			factory := jaegerProm.New()

			closer, err := cfg.InitGlobalTracer(
				golug_config.Project,
				jaegerCfg.Sampler(jaeger.NewConstSampler(true)),
				jaegerCfg.Metrics(factory),
				jaegerCfg.Logger(&tracingLogger{}),
				jaegerCfg.Reporter(jaeger.NewCompositeReporter(
					jaeger.NewLoggingReporter(&tracingLogger{}),
					jaeger.NewRemoteReporter(nil,
						jaeger.ReporterOptions.Metrics(jaeger.NewMetrics(factory, map[string]string{"lib": "jaeger"})),
						jaeger.ReporterOptions.Logger(&tracingLogger{}),
					),
				)),
			)
			xerror.Panic(err)
			xerror.Panic(dix_run.WithAfterStop(func(ctx *dix_run.AfterStopCtx) { xerror.Panic(closer.Close()) }))

		},
	}))
}
