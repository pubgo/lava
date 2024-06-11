package wss

import (
	"time"

	"github.com/pubgo/funk/log"
)

// configuration of WebsocketConnection.
type configuration struct {
	WriteDeadline       time.Duration
	ReadDeadline        *time.Duration
	PingInterval        time.Duration
	MaxInactiveDuration time.Duration
}

func defaultConfiguration() *configuration {
	return &configuration{
		WriteDeadline:       3 * time.Second,
		ReadDeadline:        nil,
		PingInterval:        30 * time.Second,
		MaxInactiveDuration: 1 * time.Minute,
	}
}

func (c *configuration) ensureValidated() {
	if c.MaxInactiveDuration <= 2*c.PingInterval {
		log.Warn().Msgf("websocket max_inactive_duration <= 2 * ping_interval, max_inactive_duration change to 3 * ping_interval")
		c.MaxInactiveDuration = 3 * c.PingInterval
	}
}

// Option is the function signature that applies configurable option for UILoadWebsocket.
type Option func(*configuration)

// ReadDeadline for how long it should wait until reading a message. Default infinite.
func ReadDeadline(d time.Duration) Option {
	return func(c *configuration) {
		c.ReadDeadline = &d
	}
}

// WriteDeadline for how long it should try before giving up writing a message. Default 3s.
func WriteDeadline(d time.Duration) Option {
	return func(c *configuration) {
		c.WriteDeadline = d
	}
}

// MaxInactiveDuration means the max time a connection can be 'quiet' before server will deem it as inactive and dead.
// Idea is that pongs and client-based pings would suffice. Default 1m.
func MaxInactiveDuration(d time.Duration) Option {
	return func(c *configuration) {
		c.MaxInactiveDuration = d
	}
}

// PingInterval sets frequency to write ping messages. Default 30s.
func PingInterval(d time.Duration) Option {
	return func(c *configuration) {
		c.PingInterval = d
	}
}
