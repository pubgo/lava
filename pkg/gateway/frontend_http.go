package gateway

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pubgo/lava/v2/pkg/httputil"
)

type httpFrontend struct {
	mux *Mux
}

func newHTTPFrontend(mux *Mux) *httpFrontend {
	return &httpFrontend{mux: mux}
}

func (f *httpFrontend) handle(ctx fiber.Ctx) error {
	webTyp, webEnc, isGRPCWeb, err := f.prepareGRPCWeb(ctx)
	if err != nil {
		return f.writeHTTPError(ctx, nil, err)
	}

	match, mth, params, err := f.match(ctx)
	if err != nil {
		var ww *fiberWebWriter
		if isGRPCWeb {
			ww = newFiberWebWriter(ctx, webTyp, webEnc)
			defer ww.flushWithTrailer()
		}
		return f.writeHTTPError(ctx, ww, err)
	}

	op := operationFromMethod(mth)
	if op.StreamDesc != nil && op.StreamDesc.ClientStreams {
		var ww *fiberWebWriter
		if isGRPCWeb {
			ww = newFiberWebWriter(ctx, webTyp, webEnc)
			defer ww.flushWithTrailer()
		}
		return f.writeHTTPError(ctx, ww, status.Error(codes.Unimplemented,
			"HTTP/gRPC-Web frontend does not support client-streaming or bidi RPCs; use WebSocket or native gRPC"))
	}

	ctx.Set(httputil.HeaderXRequestVersion, version.Version())
	ctx.Set(httputil.HeaderXRequestOperation, match.Operation)

	// Server-streams must use Fiber's stream writer so each SendMsg Flush reaches
	// the client while the handler is still running (pub/sub / long-push).
	isServerStream := op.StreamDesc != nil && op.StreamDesc.ServerStreams && !op.StreamDesc.ClientStreams
	if isServerStream {
		return f.handleServerStream(ctx, webTyp, webEnc, isGRPCWeb, match, mth, params, op)
	}

	var webWriter *fiberWebWriter
	if isGRPCWeb {
		webWriter = newFiberWebWriter(ctx, webTyp, webEnc)
		defer webWriter.flushWithTrailer()
	}

	stream := f.buildStream(ctx, mth, match, params, webWriter)
	header, trailer, err := f.mux.DispatchFrontend(stream.Context(), stream, op)
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("invoke failed")
		return f.writeHTTPError(ctx, webWriter, err)
	}

	if webWriter != nil {
		applyGRPCWebMetadata(ctx, header)
		applyGRPCWebMetadata(ctx, trailer)
		applyGRPCWebMetadata(ctx, stream.trailer)
		webWriter.ensureTrailer()
	} else {
		applyResponseMetadata(ctx, header)
		applyResponseMetadata(ctx, trailer)
		applyResponseMetadata(ctx, stream.trailer)
		ctx.Response().Header.SetContentTypeBytes(ctx.Request().Header.ContentType())
	}
	return nil
}

func (f *httpFrontend) handleServerStream(
	ctx fiber.Ctx,
	webTyp, webEnc string,
	isGRPCWeb bool,
	match *MatchOperation,
	mth *methodWrapper,
	params url.Values,
	op *Operation,
) error {
	// Snapshot request data while Fiber ctx is still valid. SendStreamWriter runs
	// later on another goroutine after the Fiber ctx may be returned to the pool.
	fctx := ctx.RequestCtx()
	stream := f.buildStream(ctx, mth, match, params, nil)
	stream.fctx = fctx
	stream.reqCT = string(ctx.Request().Header.ContentType())
	stream.reqMethod = ctx.Method()
	stream.reqBody = append([]byte(nil), ctx.Body()...)
	stream.handler = nil // do not touch pooled Fiber ctx inside the stream writer

	if isGRPCWeb {
		fctx.Response.Header.Set("Content-Type", webTyp+"+"+webEnc)
		// Negotiate compression while response headers are still mutable.
		// SendStreamWriter may flush headers on the first body write.
		stream.ensureResponseCompression()
	} else {
		fctx.Response.Header.Set(fiber.HeaderContentType, "application/x-ndjson")
	}

	return ctx.SendStreamWriter(func(bw *bufio.Writer) {
		live := &bufioHTTPFlusher{w: bw}

		var webWriter *fiberWebWriter
		if isGRPCWeb {
			webWriter = newFiberWebWriterTo(nil, fctx, webTyp, webEnc, live)
			stream.writer = webWriter
		} else {
			stream.writer = live
		}

		header, trailer, err := f.mux.DispatchFrontend(stream.Context(), stream, op)
		if err != nil {
			log.Error().
				Str("method", stream.reqMethod).
				Str("path", match.Operation).
				Msg("invoke failed")
			if webWriter != nil {
				st := status.Convert(err)
				// Prefer writer-local status so we do not race fasthttp after body bytes.
				webWriter.markErrorTrailer(st.Code(), st.Message())
				webWriter.flushWithTrailer()
			} else {
				writeNDJSONStreamError(bw, fctx, err, stream.sentHeader)
			}
			return
		}

		if webWriter != nil {
			// Only apply headers if no body has been written yet; trailers go via the writer.
			if !webWriter.wroteResp {
				applyFasthttpMetadata(fctx, header, true)
			}
			webWriter.addTrailers(trailer)
			webWriter.addTrailers(stream.trailer)
			webWriter.ensureTrailer()
			webWriter.flushWithTrailer()
			return
		}

		// HTTP/JSON: response headers must be set before the first NDJSON line.
		if !stream.sentHeader {
			applyFasthttpMetadata(fctx, header, false)
			applyFasthttpMetadata(fctx, trailer, false)
			applyFasthttpMetadata(fctx, stream.trailer, false)
		}
		_ = bw.Flush()
	})
}

// writeNDJSONStreamError emits a single JSON error object on the NDJSON stream so
// clients do not observe a silent empty 200 body when Dispatch fails.
func writeNDJSONStreamError(bw *bufio.Writer, fctx *fasthttp.RequestCtx, err error, headersSent bool) {
	st := status.Convert(err)
	if !headersSent && fctx != nil {
		fctx.Response.SetStatusCode(HTTPStatusFromCode(st.Code()))
	}
	payload, mErr := json.Marshal(map[string]any{
		"code":    uint32(st.Code()),
		"message": st.Message(),
	})
	if mErr == nil {
		_, _ = bw.Write(payload)
		_, _ = bw.Write([]byte("\n"))
	}
	_ = bw.Flush()
}

// bufioHTTPFlusher adapts *bufio.Writer to http.Flusher for per-message streaming.
type bufioHTTPFlusher struct {
	w *bufio.Writer
}

func (b *bufioHTTPFlusher) Write(p []byte) (int, error) { return b.w.Write(p) }
func (b *bufioHTTPFlusher) Flush()                      { _ = b.w.Flush() }

// writeHTTPError maps a gRPC status to the HTTP/gRPC-Web response surface.
// Plain HTTP/JSON gets an HTTP status from HTTPStatusFromCode; gRPC-Web gets
// grpc-status / grpc-message headers that flushWithTrailer emits as a trailer frame.
func (f *httpFrontend) writeHTTPError(ctx fiber.Ctx, ww *fiberWebWriter, err error) error {
	if err == nil {
		return nil
	}
	st := status.Convert(err)
	if ww != nil {
		ww.markErrorTrailer(st.Code(), st.Message())
		return nil
	}

	httpStatus := HTTPStatusFromCode(st.Code())
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	payload, mErr := json.Marshal(map[string]any{
		"code":    uint32(st.Code()),
		"message": st.Message(),
	})
	if mErr != nil {
		return ctx.Status(httpStatus).SendString(st.Message())
	}
	return ctx.Status(httpStatus).Send(payload)
}

// encodeGRPCMessage percent-encodes a grpc-message header value per gRPC-Web.
func encodeGRPCMessage(msg string) string {
	return strings.ReplaceAll(url.PathEscape(msg), "+", "%20")
}

func (f *httpFrontend) prepareGRPCWeb(ctx fiber.Ctx) (typ, enc string, ok bool, err error) {
	ct := string(ctx.Request().Header.ContentType())
	typ, enc, ok = isWebRequestFromContentType(ct, ctx.Method())
	if !ok {
		return "", "", false, nil
	}

	if strings.EqualFold(ctx.Get("Upgrade"), "websocket") {
		return "", "", false, fiber.NewError(fiber.StatusUpgradeRequired, "websocket requests must use the gateway WebSocket server (Mux.WebSocketHandler on net/http)")
	}

	ctx.Request().Header.SetContentType(grpcBase + "+" + enc)
	if typ == grpcWebText {
		if err = decodeGRPCWebTextBody(ctx); err != nil {
			return "", "", false, err
		}
	}

	return typ, enc, true, nil
}

func decodeGRPCWebTextBody(ctx fiber.Ctx) error {
	inputStream := ctx.Request().BodyStream()
	if inputStream != nil {
		body := base64.NewDecoder(base64.StdEncoding, inputStream)
		rc := &readCloser{Reader: body, Closer: io.NopCloser(nil)}
		ctx.Request().SetBodyStream(rc, -1)
		return nil
	}

	originBody := ctx.Body()
	if len(originBody) == 0 {
		return nil
	}

	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(originBody)))
	n, err := base64.StdEncoding.Decode(dbuf, originBody)
	if err != nil {
		log.Err(err).
			Stack().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("base64 decode failed")
		return errors.Errorf("base64 decode failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
	}
	ctx.Request().SetBody(dbuf[:n])
	return nil
}

func (f *httpFrontend) match(ctx fiber.Ctx) (*MatchOperation, *methodWrapper, url.Values, error) {
	matchOperation, err := f.mux.routerTree.Match(ctx.Method(), string(ctx.Request().URI().Path()))
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("match operation failed")
		return nil, nil, nil, status.Errorf(codes.NotFound, "match operation failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
	}

	values := mergePathAndQuery(ctx, matchOperation)

	mth := f.mux.findMethod(matchOperation.Operation)
	if mth == nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("method operation not found")
		return nil, nil, nil, status.Errorf(codes.NotFound, "method operation not found, method=%s", matchOperation.Operation)
	}

	return matchOperation, mth, values, nil
}

func mergePathAndQuery(ctx fiber.Ctx, match *MatchOperation) url.Values {
	values := make(url.Values)
	for _, v := range match.Vars {
		values.Set(strings.Join(v.Fields, "."), v.Value)
	}
	for k, v := range ctx.Queries() {
		values.Set(k, v)
	}
	return values
}

func (f *httpFrontend) buildStream(
	ctx fiber.Ctx,
	mth *methodWrapper,
	match *MatchOperation,
	params url.Values,
	writer *fiberWebWriter,
) *streamHTTP {
	h := make(http.Header, len(ctx.GetReqHeaders()))
	for k, v := range ctx.GetReqHeaders() {
		h[k] = append([]string(nil), v...)
	}
	reqCtx, _ := newIncomingContext(ctx.Context(), h)

	stream := &streamHTTP{
		handler: ctx,
		ctx:     reqCtx,
		method:  mth,
		params:  params,
		path:    match,
	}
	if writer != nil {
		stream.writer = writer
	}
	return stream
}
