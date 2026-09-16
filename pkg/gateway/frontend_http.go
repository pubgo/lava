package gateway

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
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
	webWriter, handled, err := f.prepareGRPCWeb(ctx)
	if err != nil {
		return f.writeHTTPError(ctx, nil, err)
	}
	if handled {
		defer webWriter.flushWithTrailer()
	}

	match, mth, params, err := f.match(ctx)
	if err != nil {
		return f.writeHTTPError(ctx, webWriter, err)
	}

	op := operationFromMethod(mth)
	if op.StreamDesc != nil && op.StreamDesc.ClientStreams {
		return f.writeHTTPError(ctx, webWriter, status.Error(codes.Unimplemented,
			"HTTP/gRPC-Web frontend does not support client-streaming or bidi RPCs; use WebSocket or native gRPC"))
	}

	stream := f.buildStream(ctx, mth, match, params, webWriter)

	ctx.Set(httputil.HeaderXRequestVersion, version.Version())
	ctx.Set(httputil.HeaderXRequestOperation, match.Operation)

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

// writeHTTPError maps a gRPC status to the HTTP/gRPC-Web response surface.
// Plain HTTP/JSON gets an HTTP status from HTTPStatusFromCode; gRPC-Web gets
// grpc-status / grpc-message headers that flushWithTrailer emits as a trailer frame.
func (f *httpFrontend) writeHTTPError(ctx fiber.Ctx, ww *fiberWebWriter, err error) error {
	if err == nil {
		return nil
	}
	st := status.Convert(err)
	if ww != nil {
		ctx.Response().Header.Set("Grpc-Status", strconv.FormatUint(uint64(st.Code()), 10))
		ctx.Response().Header.Set("Grpc-Message", encodeGRPCMessage(st.Message()))
		ww.markErrorTrailer()
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

func (f *httpFrontend) prepareGRPCWeb(ctx fiber.Ctx) (ww *fiberWebWriter, handled bool, err error) {
	ct := string(ctx.Request().Header.ContentType())
	typ, enc, ok := isWebRequestFromContentType(ct, ctx.Method())
	if !ok {
		return nil, false, nil
	}

	if strings.EqualFold(ctx.Get("Upgrade"), "websocket") {
		return nil, false, fiber.NewError(fiber.StatusUpgradeRequired, "websocket requests must use the gateway WebSocket server (Mux.WebSocketHandler on net/http)")
	}

	ctx.Request().Header.SetContentType(grpcBase + "+" + enc)
	if typ == grpcWebText {
		if err = decodeGRPCWebTextBody(ctx); err != nil {
			return nil, false, err
		}
	}

	return newFiberWebWriter(ctx, typ, enc), true, nil
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
