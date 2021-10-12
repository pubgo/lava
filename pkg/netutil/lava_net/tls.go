package lava_net

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	gnet "github.com/fatedier/golib/net"
)

var (
	FRPTLSHeadByte = 0x17
)

func WrapTLSClientConn(c net.Conn, tlsConfig *tls.Config) (out net.Conn) {
	c.Write([]byte{byte(FRPTLSHeadByte)})
	out = tls.Client(c, tlsConfig)
	return
}

func CheckAndEnableTLSServerConnWithTimeout(c net.Conn, tlsConfig *tls.Config, tlsOnly bool, timeout time.Duration) (out net.Conn, err error) {
	sc, r := gnet.NewSharedConnSize(c, 2)
	buf := make([]byte, 1)
	var n int
	c.SetReadDeadline(time.Now().Add(timeout))
	n, err = r.Read(buf)
	c.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}

	if n == 1 && int(buf[0]) == FRPTLSHeadByte {
		out = tls.Server(c, tlsConfig)
	} else {
		if tlsOnly {
			err = fmt.Errorf("non-TLS connection received on a TlsOnly server")
			return
		}
		out = sc
	}
	return
}
