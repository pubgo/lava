package wss

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pubgo/funk/log"
)

// ConstructorArgs for WebsocketConnection
type ConstructorArgs struct {
	ID           string
	Node         string
	AuthIdentity string
}

// New will return a new Websocket connection that would bridge to a message queue. For now, underlying connection is stemmed from gorilla.websocket
func New(args ConstructorArgs, conn *websocket.Conn, opts ...Option) Connection {
	cfg := defaultConfiguration()
	for _, applyOption := range opts {
		applyOption(cfg)
	}
	cfg.ensureValidated()

	return &connection{
		ConstructorArgs: args,
		Conn:            conn,
		configuration:   cfg,
		closeLock:       sync.Mutex{},
		writeLock:       sync.Mutex{},
		isClosed:        false,
		stopConnection:  func() {},
	}
}

type connection struct {
	ConstructorArgs
	*websocket.Conn
	*configuration
	closeLock      sync.Mutex
	writeLock      sync.Mutex
	isClosed       bool
	stopConnection func()
	log            log.Logger
}

func (c *connection) Write(msg []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	err := c.Conn.SetWriteDeadline(time.Now().Add(c.WriteDeadline))
	if err != nil {
		return err
	}
	return c.Conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *connection) ID() string {
	return c.ConstructorArgs.ID
}

func (c *connection) Node() string {
	return c.ConstructorArgs.Node
}

func (c *connection) AuthIdentity() string {
	return c.ConstructorArgs.AuthIdentity
}

func (c *connection) Close() error {
	var err error
	c.closeLock.Lock()
	defer func() {
		if err == nil && !c.isClosed {
			c.isClosed = true
		}
		c.closeLock.Unlock()
	}()
	if c.isClosed {
		return nil
	}
	c.stopConnection()
	return c.Conn.Close()

}

func (c *connection) Start(ctx context.Context, readMsgHandler func(int, []byte)) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	c.stopConnection = cancel
	lastMsgCh := make(chan struct{}, 1)
	go c.startMonitorForZombieConns(ctxWithCancel, lastMsgCh)
	c.Conn.SetPongHandler(func(_ string) error {
		c.log.Debug().Msgf("received pong for conn %v", c.ID())
		c.notifyLatestMessage(lastMsgCh)
		return nil
	})
	c.Conn.SetPingHandler(func(message string) error {
		c.log.Debug().Msgf("received ping for conn %v", c.ID())
		c.notifyLatestMessage(lastMsgCh)
		err := c.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(c.WriteDeadline))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	})

	for {
		select {
		case <-ctxWithCancel.Done():
			return nil
		default:
		}
		var readDeadline time.Time
		if c.ReadDeadline != nil {
			readDeadline = time.Now().Add(*c.ReadDeadline)
		}
		err := c.Conn.SetReadDeadline(readDeadline)
		if err != nil {
			return err
		}
		msgType, bytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				c.log.Err(err).Msgf("failed to read msg for %v", c.ID())
				return nil
			}
			if websocket.IsUnexpectedCloseError(err) || isClosedConnectionReadErr(err) {
				c.log.Err(err).Msgf("failed to read msg for %v", c.ID())
				return nil
			}
			c.log.Err(err).Msgf("failed to read msg for %v", c.ID())
			return nil
		}
		c.notifyLatestMessage(lastMsgCh)
		readMsgHandler(msgType, bytes)
	}
}

func (c *connection) startMonitorForZombieConns(ctx context.Context, anyMsgReceived <-chan struct{}) {
	inactiveCheckTicker := time.NewTicker(c.MaxInactiveDuration)
	defer inactiveCheckTicker.Stop()
	lastMsgReceived := time.Now()

	pingTicker := time.NewTicker(c.PingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			timeSinceLastMsg := time.Now().Sub(lastMsgReceived)
			if timeSinceLastMsg >= c.PingInterval {
				err := c.writePing([]byte(""))
				if err != nil {
					c.log.Err(err).Msgf("failed to write ping msg for conn %v", c.ID())
				}
			} else {
				c.log.Debug().Msgf("conn %v has received msg before send next ping", c.ID())
			}
			continue
		case <-inactiveCheckTicker.C:
			timeSinceLastMsg := time.Now().Sub(lastMsgReceived)
			if timeSinceLastMsg > c.MaxInactiveDuration {
				c.log.Warn().Msgf("killing connection %v because haven't received any msg [ping, pong, application] in %v seconds",
					c.ID(), timeSinceLastMsg.Seconds())
				if err := c.Close(); err != nil {
					c.log.Err(err).Msgf("failed to properly close %v", c.ID())
				}
				return
			}
			continue

		case <-anyMsgReceived:
			lastMsgReceived = time.Now()
			pingTicker.Reset(c.PingInterval)
			continue
		}
	}
}

func (c *connection) writePing(msg []byte) error {
	return c.Conn.WriteControl(websocket.PingMessage, msg, time.Now().Add(c.WriteDeadline))
}

func (c *connection) notifyLatestMessage(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
		c.log.Warn().Msgf("notify latest msg channel blocked for connection %v", c.ID())
	}
}

func isClosedConnectionReadErr(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
