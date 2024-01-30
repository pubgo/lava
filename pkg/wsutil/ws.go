package wsutil

import (
	"errors"
	"io"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
)

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

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

	// WriteBufferPool is a pool of buffers for write operations. If the value
	// is not set, then write buffers are allocated to the connection for the
	// lifetime of the connection.
	//
	// A pool is most useful when the application has a modest volume of writes
	// across a large number of connections.
	//
	// Applications should use a single pool for each unique value of
	// WriteBufferSize.
	WriteBufferPool websocket.BufferPool

	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool
}

// New returns a new `handler func(*Conn)` that upgrades a client to the
// websocket protocol, you can pass an optional config.
func New(ctx *fiber.Ctx, config ...Config) (c *websocket.Conn, err error) {
	// Init config
	var cfg Config
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

	var wsUp = websocket.FastHTTPUpgrader{
		HandshakeTimeout:  cfg.HandshakeTimeout,
		Subprotocols:      cfg.Subprotocols,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		EnableCompression: cfg.EnableCompression,
		WriteBufferPool:   cfg.WriteBufferPool,
		CheckOrigin: func(requestCtx *fasthttp.RequestCtx) bool {
			if cfg.Origins[0] == "*" {
				return true
			}
			origin := utils.UnsafeString(requestCtx.Request.Header.Peek("Origin"))
			for i := range cfg.Origins {
				if cfg.Origins[i] == origin {
					return true
				}
			}
			return false
		}}

	err = wsUp.Upgrade(ctx.Context(), func(conn *websocket.Conn) { c = conn })
	if err != nil { // Upgrading required
		return nil, fiber.ErrUpgradeRequired
	}

	return
}

// Close codes defined in RFC 6455, section 11.7.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

var (
	// ErrBadHandshake is returned when the server response to opening handshake is
	// invalid.
	ErrBadHandshake = errors.New("websocket: bad handshake")
	// ErrCloseSent is returned when the application writes a message to the
	// connection after sending a close message.
	ErrCloseSent = errors.New("websocket: close sent")
	// ErrReadLimit is returned when reading a message that is larger than the
	// read limit set for the connection.
	ErrReadLimit = errors.New("websocket: read limit exceeded")
)

// FormatCloseMessage formats closeCode and text as a WebSocket close message.
// An empty message is returned for code CloseNoStatusReceived.
func FormatCloseMessage(closeCode int, text string) []byte {
	return websocket.FormatCloseMessage(closeCode, text)
}

// IsCloseError returns boolean indicating whether the error is a *CloseError
// with one of the specified codes.
func IsCloseError(err error, codes ...int) bool {
	return websocket.IsCloseError(err, codes...)
}

// IsUnexpectedCloseError returns boolean indicating whether the error is a
// *CloseError with a code not in the list of expected codes.
func IsUnexpectedCloseError(err error, expectedCodes ...int) bool {
	return websocket.IsUnexpectedCloseError(err, expectedCodes...)
}

// IsWebSocketUpgrade returns true if the client requested upgrade to the
// WebSocket protocol.
func IsWebSocketUpgrade(c *fiber.Ctx) bool {
	return websocket.FastHTTPIsWebSocketUpgrade(c.Context())
}

// JoinMessages concatenates received messages to create a single io.Reader.
// The string term is appended to each message. The returned reader does not
// support concurrent calls to the Read method.
func JoinMessages(c *websocket.Conn, term string) io.Reader {
	return websocket.JoinMessages(c, term)
}
