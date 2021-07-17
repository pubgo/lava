package tracing

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	KeyErrorMessage        = "err_msg"
	KeyContextErrorMessage = "ctx_err_msg"
)

// SetIfErr add error info and flag for error
func SetIfErr(span opentracing.Span, err error) {
	if span == nil || err == nil {
		return
	}

	ext.Error.Set(span, true)
	span.SetTag(KeyErrorMessage, err.Error())
}

// SetIfCtxErr record error
func SetIfCtxErr(span opentracing.Span, ctx context.Context) {
	if span == nil || ctx == nil {
		return
	}

	if err := ctx.Err(); err != nil {
		ext.Error.Set(span, true)
		span.SetTag(KeyContextErrorMessage, err.Error())
	}
}

// InjectHeaders injects the outbound HTTP request with the given span's context to ensure
// correct propagation of span context throughout the trace.
func InjectHeaders(span opentracing.Span, request *http.Request) error {
	return span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header))
}

func CreateSpan(r *http.Request, name string) opentracing.Span {
	globalTracer := opentracing.GlobalTracer()

	// If headers contain trace data, create child span from parent; else, create root span
	var span opentracing.Span
	if globalTracer != nil {
		spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			span = globalTracer.StartSpan(name)
		} else {
			span = globalTracer.StartSpan(name, ext.RPCServerOption(spanCtx))
		}
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())
	}

	return span // caller must defer span.finish()
}

// Extract extracts the inbound HTTP request to obtain the parent span's context to ensure
// correct propagation of span context throughout the trace.
func Extract(tracer opentracing.Tracer, r *http.Request) (opentracing.SpanContext, error) {
	return tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header),
	)
}

// RequestFunc is a middleware function for outgoing HTTP requests.
type RequestFunc func(req *http.Request) *http.Request

// ToHTTPRequest returns a RequestFunc that injects an OpenTracing Span found in
// context into the HTTP Headers. If no such Span can be found, the RequestFunc
// is a noop.
func ToHTTPRequest(tracer opentracing.Tracer) RequestFunc {
	return func(req *http.Request) *http.Request {
		// Retrieve the Span from context.
		if span := opentracing.SpanFromContext(req.Context()); span != nil {

			// We are going to use this span in a client request, so mark as such.
			ext.SpanKindRPCClient.Set(span)

			// Add some standard OpenTracing tags, useful in an HTTP request.
			ext.HTTPMethod.Set(span, req.Method)
			span.SetTag(http.MethodPost, req.URL.Host)
			span.SetTag("path", req.URL.Path)
			ext.HTTPUrl.Set(
				span,
				fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path),
			)

			// Add information on the peer service we're about to contact.
			if host, portString, err := net.SplitHostPort(req.URL.Host); err == nil {
				ext.PeerHostname.Set(span, host)
				if port, err := strconv.Atoi(portString); err != nil {
					ext.PeerPort.Set(span, uint16(port))
				}
			} else {
				ext.PeerHostname.Set(span, req.URL.Host)
			}

			// Inject the Span context into the outgoing HTTP Request.
			if err := tracer.Inject(
				span.Context(),
				opentracing.TextMap,
				opentracing.HTTPHeadersCarrier(req.Header),
			); err != nil {
				fmt.Printf("error encountered while trying to inject span: %+v\n", err)
			}
		}
		return req
	}
}

// HandlerFunc is a middleware function for incoming HTTP requests.
type HandlerFunc func(next http.Handler) http.Handler

// FromHTTPRequest returns a Middleware HandlerFunc that tries to join with an
// OpenTracing trace found in the HTTP request headers and starts a new Span
// called `operationName`. If no trace could be found in the HTTP request
// headers, the Span will be a trace root. The Span is incorporated in the
// HTTP Context object and can be retrieved with
// opentracing.SpanFromContext(ctx).
func FromHTTPRequest(tracer opentracing.Tracer, operationName string) HandlerFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Try to join to a trace propagated in `req`.
			wireContext, err := tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header),
			)
			if err != nil {
				fmt.Printf("error encountered while trying to extract span: %+v\n", err)
			}

			//if err != nil && err != opentracing.ErrSpanContextNotFound {
			//	logger.Log("err", err)
			//}

			// create span
			span := tracer.StartSpan(operationName, ext.RPCServerOption(wireContext))
			defer span.Finish()

			// store span in context
			ctx := opentracing.ContextWithSpan(req.Context(), span)

			// update request context to include our new span
			req = req.WithContext(ctx)

			// next middleware or actual request handler
			next.ServeHTTP(w, req)
		})
	}
}
