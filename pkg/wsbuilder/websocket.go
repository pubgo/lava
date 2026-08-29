package wsbuilder

import (
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/valyala/fasthttp"
)

// WsCfg ...
type WsCfg struct {
	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration

	// Subprotocols specifies the client's requested subprotocols.
	Subprotocols []string

	// Allowed Origin's based on the Origin header, this validate the request origin to
	// prevent cross-site request forgery. Everything is allowed if left empty.
	Origins []string

	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then a useful default size is used. The I/O buffer sizes
	// do not limit the size of the messages that can be sent or received.
	ReadBufferSize, WriteBufferSize int

	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool
}

// NewWs returns a new `handler func(*Conn)` that upgrades a client to the
// websocket protocol, you can pass an optional config.
func NewWs(handler func(fiber.Ctx, *Conn), config ...WsCfg) fiber.Handler {
	// Init config
	var cfg WsCfg
	if len(config) > 0 {
		cfg = config[0]
	}
	if len(cfg.Origins) == 0 {
		cfg.Origins = []string{"*"}
	}
	if cfg.ReadBufferSize == 0 {
		cfg.ReadBufferSize = 1024
	}
	if cfg.WriteBufferSize == 0 {
		cfg.WriteBufferSize = 1024
	}
	upgrader := websocket.FastHTTPUpgrader{
		HandshakeTimeout:  cfg.HandshakeTimeout,
		Subprotocols:      cfg.Subprotocols,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		EnableCompression: cfg.EnableCompression,
		CheckOrigin: func(fctx *fasthttp.RequestCtx) bool {
			if cfg.Origins[0] == "*" {
				return true
			}
			origin := utils.UnsafeString(fctx.Request.Header.Peek("Origin"))
			for i := range cfg.Origins {
				if cfg.Origins[i] == origin {
					return true
				}
			}
			return false
		},
	}
	return func(c fiber.Ctx) error {
		conn := acquireConn()
		// locals
		c.RequestCtx().VisitUserValues(func(key []byte, value any) {
			conn.locals[string(key)] = value
		})

		// params
		params := c.Route().Params
		for i := 0; i < len(params); i++ {
			conn.params[utils.CopyString(params[i])] = utils.CopyString(c.Params(params[i]))
		}

		// queries
		for key, value := range c.RequestCtx().QueryArgs().All() {
			conn.queries[string(key)] = string(value)
		}

		// cookies
		for key, value := range c.Request().Header.All() {
			conn.cookies[string(key)] = string(value)
		}

		if err := upgrader.Upgrade(c.RequestCtx(), func(fconn *websocket.Conn) {
			conn.Conn = fconn
			defer releaseConn(conn)
			handler(c, conn)
		}); err != nil { // Upgrading required
			return fiber.ErrUpgradeRequired
		}

		return nil
	}
}

// Conn https://godoc.org/github.com/gorilla/websocket#pkg-index
type Conn struct {
	*websocket.Conn
	locals  map[string]any
	params  map[string]string
	cookies map[string]string
	queries map[string]string
}

// Conn pool
var poolConn = sync.Pool{
	New: func() any {
		return new(Conn)
	},
}

// Acquire Conn from pool
func acquireConn() *Conn {
	conn := poolConn.Get().(*Conn)
	conn.locals = make(map[string]any)
	conn.params = make(map[string]string)
	conn.queries = make(map[string]string)
	conn.cookies = make(map[string]string)
	return conn
}

// Return Conn to pool
func releaseConn(conn *Conn) {
	conn.Conn = nil
	poolConn.Put(conn)
}

// Locals makes it possible to pass interface{} values under string keys scoped to the request
// and therefore available to all following routes that match the request.
func (conn *Conn) Locals(key string) any {
	return conn.locals[key]
}

// Params is used to get the route parameters.
// Defaults to empty string "" if the param doesn't exist.
// If a default value is given, it will return that value if the param doesn't exist.
func (conn *Conn) Params(key string, defaultValue ...string) string {
	v, ok := conn.params[key]
	if !ok && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v
}

// Query returns the query string parameter in the url.
// Defaults to empty string "" if the query doesn't exist.
// If a default value is given, it will return that value if the query doesn't exist.
func (conn *Conn) Query(key string, defaultValue ...string) string {
	v, ok := conn.queries[key]
	if !ok && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v
}

// Cookies is used for getting a cookie value by key
// Defaults to empty string "" if the cookie doesn't exist.
// If a default value is given, it will return that value if the cookie doesn't exist.
func (conn *Conn) Cookies(key string, defaultValue ...string) string {
	v, ok := conn.cookies[key]
	if !ok && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return v
}
