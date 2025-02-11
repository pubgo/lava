package wsproxy

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/internal/logutil"
	"golang.org/x/net/context"
)

const (
	// Time allowed to read write a message to the peer.
	timeWait = 15 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 64
)

var (
	pingPayload = []byte("ping")
	pongPayload = []byte("pong")
)

// MethodOverrideParam defines the special URL parameter that is translated into the subsequent proxied streaming http request's method.
//
// Deprecated: it is preferable to use the Options parameters to WebSocketProxy to supply parameters.
var MethodOverrideParam = "method"

// TokenCookieName defines the cookie name that is translated to an 'Authorization: Bearer' header in the streaming http request's headers.
//
// Deprecated: it is preferable to use the Options parameters to WebSocketProxy to supply parameters.
var TokenCookieName = "token"

// RequestMutatorFunc can supply an alternate outgoing request.
type RequestMutatorFunc func(incoming, outgoing *http.Request) *http.Request

// Proxy provides websocket transport upgrade to compatible endpoints.
type Proxy struct {
	h                   http.Handler
	methodOverrideParam string
	tokenCookieName     string
	requestMutator      RequestMutatorFunc
	enablePingPong      bool
	timeWait            time.Duration
	ReadLimit           int64
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !websocket.IsWebSocketUpgrade(r) {
		p.h.ServeHTTP(w, r)
		return
	}
	p.proxy(w, r)
}

// Option allows customization of the proxy.
type Option func(*Proxy)

// WithMethodParamOverride allows specification of the special http parameter that is used in the proxied streaming request.
func WithMethodParamOverride(param string) Option {
	return func(p *Proxy) {
		p.methodOverrideParam = param
	}
}

// WithTokenCookieName allows specification of the cookie that is supplied as an upstream 'Authorization: Bearer' http header.
func WithTokenCookieName(param string) Option {
	return func(p *Proxy) {
		p.tokenCookieName = param
	}
}

// WithRequestMutator allows a custom RequestMutatorFunc to be supplied.
func WithRequestMutator(fn RequestMutatorFunc) Option {
	return func(p *Proxy) {
		p.requestMutator = fn
	}
}

func WithReadLimit(limit int64) Option {
	return func(p *Proxy) {
		p.ReadLimit = limit
	}
}

func WithPingPong(b bool) Option {
	return func(p *Proxy) {
		p.enablePingPong = b
	}
}

func WithTimeWait(t int32) Option {
	return func(p *Proxy) {
		p.timeWait = time.Second * time.Duration(t)
	}
}

// WebsocketProxy attempts to expose the underlying handler as a bidi websocket stream with newline-delimited
// JSON as the content encoding.
//
// The HTTP Authorization header is either populated from the Sec-Websocket-Protocol field or by a cookie.
// The cookie name is specified by the TokenCookieName value.
//
// example:
//
//	Sec-Websocket-Protocol: Bearer, foobar
//
// is converted to:
//
//	Authorization: Bearer foobar
//
// Method can be overwritten with the MethodOverrideParam get parameter in the requested URL
func WebsocketProxy(h http.Handler, opts ...Option) http.Handler {
	p := &Proxy{
		h:                   h,
		methodOverrideParam: MethodOverrideParam,
		tokenCookieName:     TokenCookieName,
	}
	for _, o := range opts {
		o(p)
	}

	if p.ReadLimit < maxMessageSize {
		p.ReadLimit = maxMessageSize
	}

	if p.ReadLimit > 1024*1024 {
		p.ReadLimit = 1024 * 1024
	}
	log.Debug().Any("read_limit", p.ReadLimit).Msg("read limit")
	return p
}

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func isClosedConnError(err error) bool {
	str := err.Error()
	if strings.Contains(str, "use of closed network connection") {
		return true
	}
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func (p *Proxy) proxy(w http.ResponseWriter, r *http.Request) {
	var responseHeader http.Header
	if r.Header.Get("Sec-WebSocket-Protocol") != "" {
		responseHeader = http.Header{
			"Sec-WebSocket-Protocol": []string{r.Header.Get("Sec-WebSocket-Protocol")},
		}
	}

	if p.timeWait == 0 {
		p.timeWait = timeWait
	}

	conn1, err := upgrade.Upgrade(w, r, responseHeader)
	if err != nil {
		log.Warn().Err(err).Msg("error upgrading websocket")
		return
	}
	defer conn1.Close()

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	conn := WsConn{Conn: conn1, mu: &sync.Mutex{}}
	conn.SetReadLimit(maxMessageSize)
	conn.SetPingHandler(func(text string) error {
		logutil.HandlerErr(conn.SetReadDeadline(time.Now().Add(p.timeWait)))

		log.Info().Str("text", text).Msg("websocket received ping frame")
		// 不设置 write deadline
		err := conn.WriteControl(websocket.PongMessage, []byte(text), time.Time{})
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		} else if _, ok := err.(net.Error); ok {
			return nil
		}
		return err
	})

	conn.SetPongHandler(func(text string) error {
		log.Info().Str("text", text).Msg("websocket received pong frame")
		logutil.HandlerErr(conn.SetReadDeadline(time.Now().Add(p.timeWait)))
		return nil
	})

	conn.SetCloseHandler(func(code int, text string) error {
		log.Info().Any("code", code).Any("text", text).Msg("websocket received close frame")
		cancelFn()
		return nil
	})

	if p.enablePingPong {
		log.Info().Str("time_wait", p.timeWait.String()).Msg("enable ping pong")
		logutil.HandlerErr(conn.SetReadDeadline(time.Now().Add(p.timeWait)))
	}

	requestBodyR, requestBodyW := io.Pipe()
	log.Warn().Msg("backend service only supports POST requests")
	request, err := http.NewRequest(http.MethodPost, r.URL.String(), requestBodyR)
	if err != nil {
		log.Warn().Err(err).Msg("error preparing request")
		return
	}

	for k, v := range r.Header {
		for i := range v {
			request.Header.Add(k, v[i])
		}
	}

	request.Header.Set("query", r.URL.RawQuery)

	if swsp := r.Header.Get("Sec-WebSocket-Protocol"); swsp != "" {
		request.Header.Set("Authorization", strings.Replace(swsp, "Bearer, ", "Bearer ", 1))
	}

	// If token cookie is present, populate Authorization header from the cookie instead.
	if cookie, err := r.Cookie(p.tokenCookieName); err == nil {
		request.Header.Set("Authorization", "Bearer "+cookie.Value)
	}

	if m := r.URL.Query().Get(p.methodOverrideParam); m != "" {
		request.Method = m
	}

	if p.requestMutator != nil {
		request = p.requestMutator(r, request)
	}

	responseBodyR, responseBodyW := io.Pipe()
	response := newInMemoryResponseWriter(responseBodyW)
	go func() {
		<-ctx.Done()
		log.Debug().Msg("closing websocket io pipes")
		requestBodyW.CloseWithError(io.EOF)
		responseBodyW.CloseWithError(io.EOF)
		response.closed <- true
	}()

	go func() {
		defer cancelFn()
		p.h.ServeHTTP(response, request)
	}()

	defer func() {
		log.Info().Msg("close websocket ping")
	}()

	// read loop -- take messages from websocket and write to http request
	go func() {
		defer cancelFn()
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("read loop done")
				return
			default:
				log.Debug().Msg("[read] reading from socket.")
				_, payload, err := conn.ReadMessage()
				if err != nil {
					if isClosedConnError(err) {
						log.Debug().Err(err).Msg("[read] websocket closed")
						return
					}
					log.Warn().Err(err).Msg("error reading websocket message")
					return
				}

				if p.enablePingPong {
					_ = conn.SetReadDeadline(time.Now().Add(p.timeWait))
					if bytes.Equal(payload, pingPayload) {
						logutil.HandlerErr(conn.WriteMessage(websocket.TextMessage, pongPayload))
						continue
					}
				}

				log.Debug().Str("payload", string(payload)).Msg("[read] read payload")
				log.Debug().Msg("[read] writing to requestBody:")
				n, err := requestBodyW.Write(append(payload, '\n'))
				log.Debug().Msgf("[read] wrote to requestBody %d", n)
				if err != nil {
					log.Warn().Err(err).Msg("[read] error writing message to upstream http server")
					return
				}
			}
		}
	}()

	// write loop -- take messages from response and write to websocket
	scanner := bufio.NewScanner(responseBodyR)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			log.Warn().Err(scanner.Err()).Msg("[write] empty scan")
			continue
		}

		log.Debug().Str("data", scanner.Text()).Msg("[write] scanned")
		if err = conn.WriteMessage(websocket.TextMessage, scanner.Bytes()); err != nil {
			log.Warn().Err(err).Msg("[write] error writing websocket message")
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Warn().Err(err).Msg("scanner err")
	}
}

type inMemoryResponseWriter struct {
	io.Writer
	header http.Header
	code   int
	closed chan bool
}

func newInMemoryResponseWriter(w io.Writer) *inMemoryResponseWriter {
	return &inMemoryResponseWriter{
		Writer: w,
		header: http.Header{},
		closed: make(chan bool, 1),
	}
}

func (w *inMemoryResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *inMemoryResponseWriter) Header() http.Header {
	return w.header
}

func (w *inMemoryResponseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *inMemoryResponseWriter) CloseNotify() <-chan bool {
	return w.closed
}
func (w *inMemoryResponseWriter) Flush() {}

type WsConn struct {
	*websocket.Conn
	mu *sync.Mutex
}

func (ws WsConn) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.Conn.WritePreparedMessage(pm)
}

func (ws WsConn) WriteJSON(v interface{}) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.Conn.WriteJSON(v)
}

func (ws WsConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.Conn.WriteControl(messageType, data, deadline)
}

func (ws WsConn) WriteMessage(messageType int, data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.Conn.WriteMessage(messageType, data)
}
