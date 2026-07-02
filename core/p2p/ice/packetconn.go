package ice

import (
	"net"
	"sync"
	"time"

	pionice "github.com/pion/ice/v4"
)

// packetConn 将 ICE 已连接会话适配为 net.PacketConn，供 quic-go Transport 使用。
// ICE 选路完成后等价于固定对端的 UDP 通道。
type packetConn struct {
	conn   *pionice.Conn
	remote net.Addr
	local  net.Addr
	mu     sync.Mutex
	closed bool
}

// NewPacketConn 从 ICE 连接构造 PacketConn 与对端地址。
func NewPacketConn(c *pionice.Conn) (net.PacketConn, net.Addr, error) {
	if c == nil {
		return nil, nil, errNilConn
	}
	remote := c.RemoteAddr()
	if remote == nil {
		return nil, nil, errNoRemoteAddr
	}
	pc := &packetConn{
		conn:   c,
		remote: remote,
		local:  c.LocalAddr(),
	}
	return pc, remote, nil
}

func (p *packetConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, nil, net.ErrClosed
	}
	p.mu.Unlock()
	n, err = p.conn.Read(b)
	return n, p.remote, err
}

func (p *packetConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return 0, net.ErrClosed
	}
	p.mu.Unlock()
	if addr != nil && p.remote != nil && addr.String() != p.remote.String() {
		return 0, errAddrMismatch
	}
	return p.conn.Write(b)
}

func (p *packetConn) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	return p.conn.Close()
}

func (p *packetConn) LocalAddr() net.Addr  { return p.local }
func (p *packetConn) RemoteAddr() net.Addr { return p.remote }

func (p *packetConn) SetDeadline(t time.Time) error      { return p.conn.SetDeadline(t) }
func (p *packetConn) SetReadDeadline(t time.Time) error  { return p.conn.SetReadDeadline(t) }
func (p *packetConn) SetWriteDeadline(t time.Time) error { return p.conn.SetWriteDeadline(t) }
