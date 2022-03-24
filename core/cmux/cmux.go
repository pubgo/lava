package cmux

import (
	"errors"
	"net"
	"time"

	"github.com/pubgo/xerror"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/typex"
)

func DefaultCfg() *Mux {
	return &Mux{
		Addr:        "0.0.0.0",
		Port:        8080,
		ReadTimeout: time.Second * 2,
		HandleError: func(err error) bool {
			if errors.Is(err, net.ErrClosed) {
				return true
			}

			zap.L().Named("cmux").Error("cmux matcher failed", logutil.ErrField(err)...)
			return true
		},
	}
}

type Matcher = cmux.Matcher
type matchItem struct {
	matches []cmux.Matcher
	lis     chan net.Listener
}

type Mux struct {
	Addr        string
	Port        int
	ReadTimeout time.Duration
	HandleError cmux.ErrorHandler

	ln            net.Listener
	priorityQueue typex.PriorityQueue
}

func (t *Mux) Register(priority int64, matches ...Matcher) chan net.Listener {
	var item = &matchItem{matches: matches, lis: make(chan net.Listener)}
	t.priorityQueue.PushItem(&typex.PriorityQueueItem{Value: item, Priority: priority})
	return item.lis
}

func (t *Mux) Any() chan net.Listener                { return t.Register(0, cmux.Any()) }
func (t *Mux) TLS(versions ...int) chan net.Listener { return t.Register(1, cmux.TLS(versions...)) }

func (t *Mux) HTTP1() chan net.Listener     { return t.Register(20, cmux.HTTP1()) }
func (t *Mux) HTTP1Fast() chan net.Listener { return t.Register(21, cmux.HTTP1Fast()) }
func (t *Mux) HTTP1HeaderField(name, value string) chan net.Listener {
	return t.Register(22, cmux.HTTP1HeaderField(name, value))
}

func (t *Mux) HTTP1HeaderFieldPrefix(name, valuePrefix string) chan net.Listener {
	return t.Register(23, cmux.HTTP1HeaderFieldPrefix(name, valuePrefix))
}

func (t *Mux) Websocket() chan net.Listener {
	return t.Register(24, cmux.HTTP1HeaderField("Upgrade", "websocket"))
}

func (t *Mux) HTTP2() chan net.Listener { return t.Register(30, cmux.HTTP2()) }

func (t *Mux) HTTP2HeaderField(name, value string) chan net.Listener {
	return t.Register(31, cmux.HTTP2HeaderField(name, value))
}

func (t *Mux) HTTP2HeaderFieldPrefix(name, valuePrefix string) chan net.Listener {
	return t.Register(32, cmux.HTTP2HeaderFieldPrefix(name, valuePrefix))
}

func (t *Mux) Grpc() chan net.Listener {
	return t.Register(33,
		cmux.HTTP2(),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"))
}

func (t *Mux) Close() error {
	if t.ln == nil {
		return nil
	}

	return t.ln.Close()
}

func (t *Mux) Serve() error {
	tcpAddr := &net.TCPAddr{IP: net.ParseIP(t.Addr), Port: t.Port}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	xerror.PanicF(err, "net Listen failed, addr=%s port=%d", t.Addr, t.Port)

	t.ln = ln
	var c = cmux.New(ln)
	c.SetReadTimeout(t.ReadTimeout)
	c.HandleError(t.HandleError)

	for {
		var item = t.priorityQueue.PopItem()
		if item == nil {
			break
		}

		var m = item.Value.(*matchItem)
		m.lis <- c.Match(m.matches...)
	}

	return c.Serve()
}
