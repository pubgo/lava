package netutil

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"time"

	"github.com/pubgo/xerror"
	"github.com/rsocket/rsocket-go/core"
	"github.com/rsocket/rsocket-go/core/transport"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logutil"
	"github.com/pubgo/lava/pkg/typex"
)

func DefaultCfg() *Cfg {
	return &Cfg{
		Addr:        "0.0.0.0",
		Port:        8080,
		ReadTimeout: time.Second * 2,
		HandleError: func(err error) bool {
			zap.L().Named("cmux").Error("HandleError", logutil.ErrField(err)...)
			return false
		},
	}
}

type matchItem struct {
	matches []cmux.Matcher
	l       net.Listener
}

type Cfg struct {
	ch            chan struct{}
	ln            net.Listener
	Addr          string
	Port          int
	ReadTimeout   time.Duration
	HandleError   cmux.ErrorHandler
	priorityQueue typex.PriorityQueue
}

func (t *Cfg) handler(priority int64, matches ...cmux.Matcher) func() net.Listener {
	var item = &matchItem{matches: matches}
	t.priorityQueue.PushItem(&typex.PriorityQueueItem{Value: item, Priority: priority})
	return func() net.Listener {
		// 阻塞直到初始化完毕,ch被关闭
		<-t.ch
		return item.l
	}
}

func (t *Cfg) Any() func() net.Listener                { return t.handler(0, cmux.Any()) }
func (t *Cfg) TLS(versions ...int) func() net.Listener { return t.handler(1, cmux.TLS(versions...)) }

func (t *Cfg) Rsocket() func() net.Listener {
	return t.handler(10, func(reader io.Reader) bool {
		br := bufio.NewReader(&io.LimitedReader{R: reader, N: 4096})
		l, part, err := br.ReadLine()
		if err != nil || part {
			logutil.LogOrErr(zap.L(), "ReadLine", func() error { return err })
			return false
		}

		// 用于rsocket匹配
		var frame = transport.NewLengthBasedFrameDecoder(bytes.NewBuffer(l))
		data, err := frame.Read()
		if err != nil {
			logutil.LogOrErr(zap.L(), "frame.Read", func() error { return err })
			return false
		}

		var header = core.ParseFrameHeader(data)
		if header.Type().String() == "UNKNOWN" {
			return false
		}

		if log, ok := logutil.Enabled(zap.DebugLevel); ok {
			log.Debug(header.String())
		}

		return true
	})
}

func (t *Cfg) HTTP1() func() net.Listener     { return t.handler(20, cmux.HTTP1()) }
func (t *Cfg) HTTP1Fast() func() net.Listener { return t.handler(21, cmux.HTTP1Fast()) }
func (t *Cfg) HTTP1HeaderField(name, value string) func() net.Listener {
	return t.handler(22, cmux.HTTP1HeaderField(name, value))
}

func (t *Cfg) HTTP1HeaderFieldPrefix(name, valuePrefix string) func() net.Listener {
	return t.handler(23, cmux.HTTP1HeaderFieldPrefix(name, valuePrefix))
}

func (t *Cfg) Websocket() func() net.Listener {
	return t.handler(24, cmux.HTTP1HeaderField("Upgrade", "websocket"))
}

func (t *Cfg) HTTP2() func() net.Listener { return t.handler(30, cmux.HTTP2()) }

func (t *Cfg) HTTP2HeaderField(name, value string) func() net.Listener {
	return t.handler(31, cmux.HTTP2HeaderField(name, value))
}

func (t *Cfg) HTTP2HeaderFieldPrefix(name, valuePrefix string) func() net.Listener {
	return t.handler(32, cmux.HTTP2HeaderFieldPrefix(name, valuePrefix))
}

func (t *Cfg) Grpc() func() net.Listener {
	return t.handler(33,
		cmux.HTTP2(),
		cmux.HTTP2HeaderFieldPrefix("content-type", "application/grpc"))
}

func (t *Cfg) Close() error {
	if t.ln == nil {
		return nil
	}

	return t.ln.Close()
}

func (t *Cfg) Serve() error {
	tcpAddr := &net.TCPAddr{IP: net.ParseIP(t.Addr), Port: t.Port}
	tcpLn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return xerror.WrapF(err, "net.ListenTCP failed, addr=>%s, port=>%d", t.Addr, t.Port)
	}

	t.ln = tcpLn
	var c = cmux.New(tcpLn)
	c.SetReadTimeout(t.ReadTimeout)
	c.HandleError(t.HandleError)

	for {
		var item = t.priorityQueue.PopItem()
		if item == nil {
			break
		}

		var m = item.Value.(*matchItem)
		m.l = c.Match(m.matches...)
	}

	// 初始化完毕后, 关闭ch
	close(t.ch)
	return c.Serve()
}
