package cluster

import (
	"errors"
	"fmt"
	"github.com/pubgo/lava/core/cmux"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/armon/go-metrics"
	sockAddr "github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/memberlist"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logutil"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65536

	// udpRecvBufSize is a large buffer size that we attempt to set UDP
	// sockets to in order to handle a large volume of messages.
	udpRecvBufSize = 2 * 1024 * 1024
)

var _ memberlist.NodeAwareTransport = (*netTransport)(nil)

// netTransport is a Transport implementation that uses connectionless UDP for
// packet operations, and ad-hoc TCP connections for stream operations.
type netTransport struct {
	logger      *zap.Logger
	netCfg      *cmux.Mux
	packetCh    chan *memberlist.Packet
	streamCh    chan net.Conn
	wg          sync.WaitGroup
	tcpListener func() net.Listener
	udpListener *net.UDPConn
	shutdown    int32
}

// newNetTransport returns a net transport with the given configuration. On
// success all the network listeners will be created and listening.
func newNetTransport(log *zap.Logger, netCfg *cmux.Mux) *netTransport {
	defer xerror.RespExit()

	// If we reject the empty list outright we can assume that there's at
	// least one listener of each type later during operation.
	xerror.Assert(netCfg == nil || len(netCfg.Addr) == 0, "at least one bind address is required")
	xerror.Assert(log == nil, "log should not be nil")

	// Build all the TCP and UDP listeners.
	if netCfg.ReadTimeout == 0 {
		netCfg.ReadTimeout = time.Second * 2
	}

	var logger = log.Named("transport")

	netCfg.HandleError = func(err error) bool {
		logger.Error("HandleError", logutil.ErrField(err)...)
		return false
	}

	// Build out the new transport.
	t := netTransport{
		netCfg:   netCfg,
		logger:   logger,
		packetCh: make(chan *memberlist.Packet),
		streamCh: make(chan net.Conn),
	}

	t.tcpListener = netCfg.Any()

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		if err := netCfg.Serve(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, net.ErrClosed) {
			logger.Error("net close failed", logutil.ErrField(err)...)
		}
	}()

	ip := net.ParseIP(netCfg.Addr)
	udpAddr := &net.UDPAddr{IP: ip, Port: netCfg.Port}
	udpLn, err := net.ListenUDP("udp", udpAddr)
	xerror.PanicF(err, "failed to start UDP listener on %q port %d: %v", netCfg.Addr, netCfg.Port, err)

	xerror.PanicF(setUDPRecvBuf(udpLn), "failed to resize UDP buffer")

	t.udpListener = udpLn

	t.wg.Add(2)
	go t.tcpListen(t.tcpListener)
	go t.udpListen(t.udpListener)
	return &t
}

// GetAutoBindPort returns the bind port that was automatically given by the
// kernel, if a bind port of 0 was given.
func (t *netTransport) GetAutoBindPort() int { return t.netCfg.Port }

// FinalAdvertiseAddr See Transport.
func (t *netTransport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	var advertiseAddr net.IP
	var advertisePort int
	if ip != "" {
		// If they've supplied an address, use that.
		advertiseAddr = net.ParseIP(ip)
		if advertiseAddr == nil {
			return nil, 0, fmt.Errorf("Failed to parse advertise address %q", ip)
		}

		// Ensure IPv4 conversion if necessary.
		if ip4 := advertiseAddr.To4(); ip4 != nil {
			advertiseAddr = ip4
		}
		advertisePort = port
	} else {
		if t.netCfg.Addr == "0.0.0.0" {
			// Otherwise, if we're not bound to a specific IP, let's
			// use a suitable private IP address.
			var err error
			ip, err = sockAddr.GetPrivateIP()
			if err != nil {
				return nil, 0, fmt.Errorf("failed to get interface addresses: %v", err)
			}
			if ip == "" {
				return nil, 0, fmt.Errorf("no private IP address found, and explicit IP not provided")
			}

			advertiseAddr = net.ParseIP(ip)
			if advertiseAddr == nil {
				return nil, 0, fmt.Errorf("failed to parse advertise address: %q", ip)
			}
		} else {
			// Use the IP that we're bound to, based on the first
			// TCP listener, which we already ensure is there.
			advertiseAddr = t.tcpListener().Addr().(*net.TCPAddr).IP
		}

		// Use the port we are bound to.
		advertisePort = t.GetAutoBindPort()
	}

	return advertiseAddr, advertisePort, nil
}

// WriteTo See Transport.
func (t *netTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return time.Time{}, err
	}

	// We made sure there's at least one UDP listener, so just use the
	// packet sending interface on the first one. Take the time after the
	// write call comes back, which will underestimate the time a little,
	// but help account for any delays before the write occurs.
	_, err = t.udpListener.WriteTo(b, udpAddr)
	return time.Now(), err
}

// WriteToAddress See NodeAwareTransport.
func (t *netTransport) WriteToAddress(b []byte, a memberlist.Address) (time.Time, error) {
	return t.WriteTo(b, a.Addr)
}

// PacketCh See Transport.
func (t *netTransport) PacketCh() <-chan *memberlist.Packet { return t.packetCh }

// DialTimeout See Transport.
func (t *netTransport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	dialer := net.Dialer{Timeout: timeout}
	return dialer.Dial("tcp", addr)
}

func (t *netTransport) DialAddressTimeout(addr memberlist.Address, timeout time.Duration) (net.Conn, error) {
	return t.DialTimeout(addr.Addr, timeout)
}

// StreamCh See Transport.
func (t *netTransport) StreamCh() <-chan net.Conn { return t.streamCh }

// Shutdown See Transport.
func (t *netTransport) Shutdown() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&t.shutdown, 1)

	// Rip through all the connections and shut them down.
	t.udpListener.Close()

	t.tcpListener().Close()

	t.netCfg.Close()

	// Block until all the listener threads have died.
	t.wg.Wait()
	return nil
}

// tcpListen is a long running goroutine that accepts incoming TCP connections
// and hands them off to the stream channel.
func (t *netTransport) tcpListen(tcpLn func() net.Listener) {
	defer t.wg.Done()

	// baseDelay is the initial delay after an AcceptTCP() error before attempting again
	const baseDelay = 5 * time.Millisecond

	// maxDelay is the maximum delay after an AcceptTCP() error before attempting again.
	// In the case that tcpListen() is error-looping, it will delay the shutdown check.
	// Therefore, changes to maxDelay may have an effect on the latency of shutdown.
	const maxDelay = 1 * time.Second

	var loopDelay time.Duration
	for {
		conn, err := tcpLn().Accept()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			if loopDelay == 0 {
				loopDelay = baseDelay
			} else {
				loopDelay *= 2
			}

			if loopDelay > maxDelay {
				loopDelay = maxDelay
			}

			logs.S().Errorf("[ERR] memberlist: Error accepting TCP connection: %v", err)
			time.Sleep(loopDelay)
			continue
		}

		// No error, reset loop delay
		loopDelay = 0

		t.streamCh <- conn
	}
}

// udpListen is a long running goroutine that accepts incoming UDP packets and
// hands them off to the packet channel.
func (t *netTransport) udpListen(udpLn *net.UDPConn) {
	defer t.wg.Done()

	for {
		// Do a blocking read into a fresh buffer. Grab a time stamp as
		// close as possible to the I/O.
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := udpLn.ReadFrom(buf)
		ts := time.Now()
		if err != nil {
			if s := atomic.LoadInt32(&t.shutdown); s == 1 {
				break
			}

			logs.S().Errorf("[ERR] memberlist: Error reading UDP packet: %v", err)
			continue
		}

		// Check the length - it needs to have at least one byte to be a
		// proper message.
		if n < 1 {
			logs.S().Errorf("[ERR] memberlist: UDP packet too short (%d bytes) %s",
				len(buf), memberlist.LogAddress(addr))
			continue
		}

		// Ingest the packet.
		metrics.IncrCounter([]string{"memberlist", "udp", "received"}, float32(n))
		t.packetCh <- &memberlist.Packet{
			Buf:       buf[:n],
			From:      addr,
			Timestamp: ts,
		}
	}
}

// setUDPRecvBuf is used to resize the UDP receive window. The function
// attempts to set the read buffer to `udpRecvBuf` but backs off until
// the read buffer can be set.
func setUDPRecvBuf(c *net.UDPConn) error {
	size := udpRecvBufSize
	var err error
	for size > 0 {
		if err = c.SetReadBuffer(size); err == nil {
			return nil
		}
		size = size / 2
	}
	return err
}
