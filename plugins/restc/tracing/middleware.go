package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/pubgo/lug/plugins/restc"
	"github.com/pubgo/lug/tracing"

	"net/http"
)

func Middleware(opts ...interface{}) restc.Middleware {
	return func(next restc.DoFunc) restc.DoFunc {
		return func(req *restc.Request, fn func(resp *restc.Response) error) error {
			var span = tracing.FromCtx(req)
			if span == nil {
				span = tracing.NewSpan(tracing.CreateSpanFromFast(req.Request, req.URI().String()))
			}

			var header = make(http.Header)

			// span信息埋入http head
			_ = span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))

			for k, v := range header {
				for i := range v {
					req.Header.Add(k, v[i])
				}
			}

			var resp *restc.Response
			var err error
			defer func() {
				ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode()))
				tracing.SetIfErr(span, err)
				span.Finish()
			}()

			return next(req, func(response *restc.Response) error { resp = response; err = fn(response); return err })
		}
	}
}
