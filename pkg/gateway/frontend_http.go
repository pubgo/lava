package gateway

import (
	"encoding/base64"
	"io"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/lava/v2/pkg/httputil"
	"google.golang.org/grpc/metadata"
)

type httpFrontend struct {
	mux        *Mux
	dispatcher *Dispatcher
}

func newHTTPFrontend(mux *Mux) *httpFrontend {
	return &httpFrontend{
		mux:        mux,
		dispatcher: NewDispatcher(),
	}
}

func (f *httpFrontend) handle(ctx fiber.Ctx) error {
	webWriter, handled, err := f.prepareGRPCWeb(ctx)
	if err != nil {
		return err
	}
	if handled {
		defer webWriter.flushWithTrailer()
	}

	match, mth, params, err := f.match(ctx)
	if err != nil {
		return err
	}

	stream := f.buildStream(ctx, mth, match, params, webWriter)
	in := mth.inputType.New().Interface()
	if err = stream.RecvMsg(in); err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("unmarshal request failed")
		return errors.Errorf("unmarshal request failed, method=%s", match.Operation)
	}

	ctx.Set(httputil.HeaderXRequestVersion, version.Version())
	ctx.Set(httputil.HeaderXRequestOperation, match.Operation)

	header, trailer, err := f.dispatcher.Dispatch(stream.Context(), f.mux, stream, operationFromMethod(mth), in)
	if err != nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("invoke failed")
		return errors.WrapCaller(err)
	}

	applyResponseMetadata(ctx, header)
	applyResponseMetadata(ctx, trailer)
	applyResponseMetadata(ctx, stream.trailer)

	if webWriter == nil {
		ctx.Response().Header.SetContentTypeBytes(ctx.Request().Header.ContentType())
	}
	return nil
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
		return nil, nil, nil, errors.Errorf("match operation failed, method=%s path=%s", ctx.Method(), string(ctx.Request().URI().Path()))
	}

	values := mergePathAndQuery(ctx, matchOperation)

	mth := f.mux.opts.handlers[matchOperation.Operation]
	if mth == nil {
		log.Error().
			Str("method", ctx.Method()).
			Str("path", string(ctx.Request().URI().Path())).
			Msg("method operation not found")
		return nil, nil, nil, errors.Errorf("method operation not found, method=%s", matchOperation.Operation)
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
	md := metadata.MD{}
	for k, v := range ctx.GetReqHeaders() {
		md.Append(k, v...)
	}

	stream := &streamHTTP{
		handler: ctx,
		ctx:     metadata.NewIncomingContext(ctx.Context(), md),
		method:  mth,
		params:  params,
		path:    match,
	}
	if writer != nil {
		stream.writer = writer
	}
	return stream
}
