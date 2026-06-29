package gateway

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/coder/websocket"
	"github.com/pubgo/funk/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// WSOption configures the websocket frontend.
type WSOption func(*wsOptions)

type wsOptions struct {
	// originPatterns is passed to websocket.AcceptOptions.OriginPatterns.
	// Empty means same-origin only (coder/websocket default).
	originPatterns []string
	// insecureSkipVerify disables origin checks. Use with care.
	insecureSkipVerify bool
	// subprotocols advertised during the handshake.
	subprotocols []string
}

// WithWSOriginPatterns sets allowed origin patterns for the websocket handshake.
func WithWSOriginPatterns(patterns ...string) WSOption {
	return func(o *wsOptions) { o.originPatterns = patterns }
}

// WithWSInsecureSkipVerify disables websocket origin verification.
func WithWSInsecureSkipVerify() WSOption {
	return func(o *wsOptions) { o.insecureSkipVerify = true }
}

// WithWSSubprotocols sets the advertised websocket subprotocols.
func WithWSSubprotocols(subprotocols ...string) WSOption {
	return func(o *wsOptions) { o.subprotocols = subprotocols }
}

type wsFrontend struct {
	mux        *Mux
	dispatcher *Dispatcher
	opts       wsOptions
}

// WebSocketHandler returns an http.Handler that bridges websocket clients to the
// registered gRPC handlers. It uses coder/websocket, so it must be served by a
// standard net/http server (it cannot run on the fasthttp/fiber stack).
//
// The request path is interpreted as the gRPC full method name, e.g.
// "/pkg.v1.Service/Method". The wire format defaults to protojson (text frames)
// and can be switched to protobuf (binary frames) via the "?encoding=proto"
// query parameter or the "grpc-ws-proto" subprotocol.
func (m *Mux) WebSocketHandler(opts ...WSOption) http.Handler {
	f := &wsFrontend{mux: m, dispatcher: m.dispatcher}
	for _, opt := range opts {
		opt(&f.opts)
	}
	return f
}

func (f *wsFrontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	op, mth, match, params := f.resolveOperation(r)
	if op == nil {
		http.Error(w, "method operation not found: "+r.URL.Path, http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:       f.opts.subprotocols,
		OriginPatterns:     f.opts.originPatterns,
		InsecureSkipVerify: f.opts.insecureSkipVerify,
	})
	if err != nil {
		log.Err(err).Str("path", r.URL.Path).Msg("websocket accept failed")
		return
	}
	defer func() { _ = conn.CloseNow() }()

	enc := resolveWSEncoding(r, conn.Subprotocol())

	md := metadata.MD{}
	for k, vs := range r.Header {
		md.Append(k, vs...)
	}

	stream := &streamWS{
		conn:     conn,
		ctx:      metadata.NewIncomingContext(r.Context(), md),
		method:   mth,
		encoding: enc,
		path:     match,
		params:   params,
	}

	if err = f.dispatch(stream, op); err != nil {
		log.Err(err).Str("method", op.FullMethod).Msg("websocket dispatch failed")
		closeWithStatus(conn, err)
		return
	}

	closeWithStatus(conn, nil)
}

// closeWithStatus closes the websocket with a structured gRPC status carried in
// the close frame: the close code is mapped from the gRPC code, and the reason
// is a JSON object {"grpcStatus":N,"grpcMessage":"..."} so clients can recover
// the gRPC status/message instead of relying on the close code alone.
func closeWithStatus(conn *websocket.Conn, err error) {
	st := status.Convert(err)
	payload := struct {
		GRPCStatus  uint32 `json:"grpcStatus"`
		GRPCMessage string `json:"grpcMessage"`
	}{
		GRPCStatus:  uint32(st.Code()),
		GRPCMessage: st.Message(),
	}

	reason := ""
	if b, mErr := json.Marshal(payload); mErr == nil {
		reason = string(b)
	}
	// WebSocket close reason is limited to 123 bytes.
	if len(reason) > 123 {
		reason = reason[:123]
	}

	_ = conn.Close(wsCloseCode(st.Code()), reason)
}

// wsCloseCode maps a gRPC code to a WebSocket close code. OK uses the normal
// closure; all error codes use the internal-error code so existing clients that
// only inspect the close code still observe a non-normal closure.
func wsCloseCode(code codes.Code) websocket.StatusCode {
	if code == codes.OK {
		return websocket.StatusNormalClosure
	}
	return websocket.StatusInternalError
}

func (f *wsFrontend) dispatch(stream *streamWS, op *Operation) error {
	_, _, err := f.dispatcher.DispatchFrontend(stream.Context(), f.mux, stream, op)
	return err
}

func (f *wsFrontend) resolveOperation(r *http.Request) (*Operation, *methodWrapper, *MatchOperation, url.Values) {
	path := r.URL.Path

	// Direct gRPC full-method lookup: /pkg.Service/Method
	if mth := f.mux.opts.handlers[path]; mth != nil {
		return operationFromMethod(mth), mth, nil, nil
	}

	// Fall back to REST-style routing so HTTP-annotated paths also work over WS.
	if match, err := f.mux.routerTree.Match(r.Method, path); err == nil {
		if mth := f.mux.opts.handlers[match.Operation]; mth != nil {
			values := make(url.Values)
			for _, v := range match.Vars {
				values.Set(strings.Join(v.Fields, "."), v.Value)
			}
			for k, vs := range r.URL.Query() {
				for _, v := range vs {
					values.Set(k, v)
				}
			}
			return operationFromMethod(mth), mth, match, values
		}
	}

	return nil, nil, nil, nil
}

func resolveWSEncoding(r *http.Request, subprotocol string) wsEncoding {
	switch {
	case subprotocol == "grpc-ws-proto":
		return wsEncodingProto
	case subprotocol == "grpc-ws-json":
		return wsEncodingJSON
	}

	switch strings.ToLower(r.URL.Query().Get("encoding")) {
	case "proto", "protobuf", "binary":
		return wsEncodingProto
	default:
		return wsEncodingJSON
	}
}
